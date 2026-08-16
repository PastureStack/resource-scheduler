ARG SOURCE_DATE_EPOCH
ARG UBUNTU_IMAGE=ubuntu:26.04@sha256:7c2884fd32770fc6c173b78e0dc2278a2851d89f5447919edbc45475ac55dd6a
FROM ${UBUNTU_IMAGE} AS snapshot-ca-bootstrap
ADD --checksum=sha256:6077d27c6b6f8b23590cb01ff877ed8c804a67a5442cc32b5a33da10d2bd0e90 \
    https://snapshot.ubuntu.com/ubuntu/20260808T000000Z/pool/main/c/ca-certificates/ca-certificates_20260601~26.04.1_all.deb \
    /tmp/ca-certificates.deb
RUN dpkg-deb -x /tmp/ca-certificates.deb /tmp/ca-certificates && \
    cp /tmp/ca-certificates/usr/share/ca-certificates/mozilla/ISRG_Root_X1.crt /snapshot-ca.crt

FROM ${UBUNTU_IMAGE}
COPY --from=snapshot-ca-bootstrap /snapshot-ca.crt /etc/ssl/certs/ca-certificates.crt
COPY ubuntu-apt.lock /usr/share/pasturestack/manifests/ubuntu-apt.lock

ARG IMAGE_VERSION=dev
ARG IMAGE_REVISION=unknown
ARG IMAGE_CREATED=unknown
ARG SOURCE_DATE_EPOCH
ARG UBUNTU_SNAPSHOT=20260808T000000Z
ARG UBUNTU_SNAPSHOT_BASE=https://snapshot.ubuntu.com/ubuntu
ENV DEBIAN_FRONTEND=noninteractive

RUN set -eux; \
    . /usr/share/pasturestack/manifests/ubuntu-apt.lock; \
    test "${UBUNTU_SNAPSHOT}" = "${UBUNTU_APT_SNAPSHOT}"; \
    snapshot_uri="${UBUNTU_SNAPSHOT_BASE%/}/${UBUNTU_SNAPSHOT}"; \
    rm -f /etc/apt/sources.list /etc/apt/sources.list.d/*.list /etc/apt/sources.list.d/*.sources; \
    printf 'Types: deb\nURIs: %s\nSuites: resolute resolute-updates resolute-backports resolute-security\nComponents: main universe restricted multiverse\nSigned-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg\nSnapshot: no\n' \
        "${snapshot_uri}" > /etc/apt/sources.list.d/pasturestack-snapshot.sources; \
    printf 'Acquire::Retries "5";\nAcquire::http::Timeout "30";\nAcquire::https::Timeout "30";\nAcquire::https::CaInfo "/etc/ssl/certs/ca-certificates.crt";\nAcquire::https::Verify-Peer "true";\nAcquire::https::Verify-Host "true";\nAcquire::http::Pipeline-Depth "0";\nAcquire::AllowInsecureRepositories "false";\nAPT::Get::AllowUnauthenticated "false";\n' > /etc/apt/apt.conf.d/80pasturestack-retries; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        bash="${UBUNTU_APT_BASH_VERSION}" \
        ca-certificates="${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        curl="${UBUNTU_APT_CURL_VERSION}" \
        libcap2-bin="${UBUNTU_APT_LIBCAP2_BIN_VERSION}"; \
    for package_and_version in \
        "bash=${UBUNTU_APT_BASH_VERSION}" \
        "ca-certificates=${UBUNTU_APT_CA_CERTIFICATES_VERSION}" \
        "curl=${UBUNTU_APT_CURL_VERSION}" \
        "libcap2-bin=${UBUNTU_APT_LIBCAP2_BIN_VERSION}"; do \
        package="${package_and_version%%=*}"; \
        expected="${package_and_version#*=}"; \
        test "$(dpkg-query -W -f='${Version}' "${package}")" = "${expected}"; \
    done; \
    { \
        printf 'record\tname\tarchitecture\tversion\n'; \
        printf 'metadata\tschema\t-\tpasturestack.dpkg-manifest/v1\n'; \
        printf 'metadata\tubuntu_snapshot\t-\t%s\n' "${UBUNTU_SNAPSHOT}"; \
        dpkg-query -W -f='package\t${binary:Package}\t${Architecture}\t${Version}\n' | LC_ALL=C sort; \
    } > /usr/share/pasturestack/manifests/runtime-ubuntu-packages.tsv; \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*; \
    rm -f /var/log/apt/* /var/log/dpkg.log /var/log/alternatives.log /var/log/lastlog /var/log/faillog /var/cache/ldconfig/aux-cache; \
    rm -f /usr/bin/pebble

COPY bin/resource-scheduler /usr/bin/resource-scheduler
COPY package/pasturestack-entrypoint.sh package/update-pasturestack-ca /usr/bin/
COPY LICENSE COMPATIBILITY.md ORIGIN.md SECURITY.md THIRD_PARTY_NOTICES.md /licenses/
COPY LICENSES /licenses/third-party/

RUN set -eux; \
    sed -i 's/\r$//' /usr/bin/pasturestack-entrypoint.sh /usr/bin/update-pasturestack-ca; \
    chmod 0755 /usr/bin/resource-scheduler /usr/bin/pasturestack-entrypoint.sh /usr/bin/update-pasturestack-ca; \
    ln -s /usr/bin/resource-scheduler /usr/bin/scheduler; \
    groupadd --gid 10001 pasturestack; \
    useradd --uid 10001 --gid pasturestack --home-dir /home/pasturestack --no-create-home --no-log-init --shell /usr/sbin/nologin pasturestack; \
    mkdir -p /home/pasturestack; \
    chown 10001:10001 /home/pasturestack; \
    setcap cap_net_bind_service=+ep /usr/bin/resource-scheduler; \
    rm -f /var/log/lastlog /var/log/faillog

ENV HOME=/home/pasturestack
USER 10001:10001

LABEL org.opencontainers.image.title="PastureStack Resource Scheduler" \
      org.opencontainers.image.description="Resource, placement, and host-port scheduling service for PastureStack" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/PastureStack/resource-scheduler" \
      org.opencontainers.image.documentation="https://github.com/PastureStack/resource-scheduler/blob/main/README.md" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.created="${IMAGE_CREATED}"

ENTRYPOINT ["/usr/bin/pasturestack-entrypoint.sh"]
CMD ["resource-scheduler"]
