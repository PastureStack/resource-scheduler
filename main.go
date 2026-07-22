package main

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PastureStack/resource-scheduler/events"
	"github.com/PastureStack/resource-scheduler/internal/metadata"
	"github.com/PastureStack/resource-scheduler/resourcewatchers"
	"github.com/PastureStack/resource-scheduler/scheduler"
	pkgerrors "github.com/pkg/errors"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/log"
	logserver "github.com/rancher/log/server"
	"github.com/urfave/cli"
)

var VERSION = "v0.1.0-dev"

func main() {
	logserver.StartServerWithDefaults()
	metadataAddress := os.Getenv("PASTURESTACK_METADATA_ADDRESS")
	if metadataAddress == "" {
		metadataAddress = "metadata"
	}

	app := cli.NewApp()
	app.Name = "resource-scheduler"
	app.Version = VERSION
	app.Usage = "PastureStack resource and host-port scheduling service."
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "metadata-address",
			Usage: "Metadata service address",
			Value: metadataAddress,
		},
		cli.IntFlag{
			Name:  "health-check-port",
			Usage: "Port to listen on for health checks",
			Value: 80,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Resource scheduler exited: %v", err)
	}
}

func run(c *cli.Context) error {
	if debugEnabled() {
		log.SetLevelString("debug")
	}

	sleepSeconds := 1
	if rawSleep := os.Getenv("CATTLE_SCHEDULER_SLEEPTIME"); rawSleep != "" {
		value, err := strconv.Atoi(rawSleep)
		if err != nil || value < 0 {
			log.Warnf("Ignoring invalid scheduler sleep interval %q", rawSleep)
		} else {
			sleepSeconds = value
		}
	}
	resourceScheduler := scheduler.NewScheduler(sleepSeconds)
	metadataClient := metadata.NewClient(fmt.Sprintf("http://%s/2016-07-29", c.String("metadata-address")))
	resourceScheduler.SetMetadataClient(metadataClient)

	controlPlaneURL := os.Getenv("CATTLE_URL")
	accessKey := os.Getenv("CATTLE_ACCESS_KEY")
	secretKey := os.Getenv("CATTLE_SECRET_KEY")
	if controlPlaneURL == "" || accessKey == "" || secretKey == "" {
		return stderrors.New("control-plane connection environment variables are incomplete")
	}
	apiClient, err := client.NewRancherClient(&client.ClientOpts{
		Timeout:   30 * time.Second,
		Url:       controlPlaneURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if err != nil {
		return err
	}

	exit := make(chan error)
	go func() {
		err := events.ConnectToEventStream(controlPlaneURL, accessKey, secretKey, resourceScheduler)
		exit <- pkgerrors.Wrap(err, "control-plane event subscriber exited")
	}()

	go func() {
		err := resourcewatchers.WatchMetadata(metadataClient, resourceScheduler, apiClient)
		exit <- pkgerrors.Wrap(err, "metadata watcher exited")
	}()

	go func() {
		err := startHealthCheck(c.Int("health-check-port"), metadataClient, controlPlaneURL)
		exit <- pkgerrors.Wrap(err, "health-check provider exited")
	}()

	go func() {
		for {
			time.Sleep(3 * time.Minute)
			log.Info("Synchronizing scheduler state with control-plane metadata")
			for {
				ok, err := resourceScheduler.UpdateWithMetadata(true)
				if err != nil {
					log.Warnf("Error synchronizing metadata: %v", err)
					break
				}
				if !ok {
					log.Info("Delaying metadata synchronization while events are active")
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		}
	}()

	err = <-exit
	log.Errorf("Resource scheduler is exiting: %v", err)
	return err
}

func debugEnabled() bool {
	for _, name := range []string{"PASTURESTACK_DEBUG", "RANCHER_DEBUG"} {
		if strings.EqualFold(os.Getenv(name), "true") {
			return true
		}
	}
	return false
}

func startHealthCheck(listen int, metadataClient metadata.Client, controlPlaneURL string) error {
	pingURL, err := controlPlanePingURL(controlPlaneURL)
	if err != nil {
		return err
	}
	pingClient := &http.Client{Timeout: 10 * time.Second}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			response.Header().Set("Allow", "GET, HEAD")
			http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := metadataClient.GetVersion(); err != nil {
			log.Errorf("Health check could not reach metadata: %v", err)
			http.Error(response, "metadata is unavailable", http.StatusServiceUnavailable)
			return
		}
		pingResponse, err := pingClient.Get(pingURL)
		if err != nil {
			log.Errorf("Health check could not reach the control plane: %v", err)
			http.Error(response, "control plane is unavailable", http.StatusServiceUnavailable)
			return
		}
		defer pingResponse.Body.Close()
		if pingResponse.StatusCode < http.StatusOK || pingResponse.StatusCode >= http.StatusMultipleChoices {
			log.Errorf("Control-plane ping returned HTTP %d", pingResponse.StatusCode)
			http.Error(response, "control plane is unavailable", http.StatusServiceUnavailable)
			return
		}
		response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(response, "ok")
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", listen),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	log.Infof("Listening for health checks on %s/healthcheck", server.Addr)
	return server.ListenAndServe()
}

func controlPlanePingURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", stderrors.New("invalid control-plane URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", stderrors.New("control-plane URL must not contain credentials, a query, or a fragment")
	}
	path := strings.TrimRight(parsed.Path, "/")
	lastSlash := strings.LastIndex(path, "/")
	lastSegment := path
	if lastSlash >= 0 {
		lastSegment = path[lastSlash+1:]
	}
	if isAPIVersionSegment(lastSegment) {
		path = path[:lastSlash]
	}
	parsed.Path = strings.TrimRight(path, "/") + "/ping"
	parsed.RawPath = ""
	return parsed.String(), nil
}

func isAPIVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' || segment[1] < '0' || segment[1] > '9' {
		return false
	}
	for _, character := range segment[2:] {
		if (character < '0' || character > '9') && character != '-' && (character < 'a' || character > 'z') {
			return false
		}
	}
	return true
}
