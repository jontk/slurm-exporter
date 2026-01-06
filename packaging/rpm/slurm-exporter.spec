Name:           slurm-exporter
Version:        %{version}
Release:        1%{?dist}
Summary:        Prometheus exporter for SLURM workload manager metrics

License:        MIT
URL:            https://github.com/jontk/slurm-exporter
Source0:        https://github.com/jontk/slurm-exporter/archive/v%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  git
BuildRequires:  systemd-rpm-macros

Requires:       systemd
Requires(pre):  shadow-utils
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

# Disable debug package generation
%global debug_package %{nil}

# Disable automatic dependency detection for Go binaries
%global __requires_exclude_from ^%{_bindir}/.*$

%description
SLURM Exporter is a comprehensive Prometheus exporter designed specifically 
for SLURM (Simple Linux Utility for Resource Management) workload managers. 
It provides detailed metrics about jobs, nodes, partitions, users, accounts, 
fairshare policies, and quality of service (QoS) settings.

Key features:
- Comprehensive SLURM metrics collection
- Advanced job analytics and efficiency monitoring
- High performance with intelligent caching
- Production-ready with security hardening
- Kubernetes and container native support

%prep
%autosetup -n %{name}-%{version}

%build
# Set Go build environment
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=%{_arch}

# Map RPM arch to Go arch
%ifarch x86_64
export GOARCH=amd64
%endif
%ifarch aarch64
export GOARCH=arm64
%endif
%ifarch i386 i686
export GOARCH=386
%endif

# Build the binary
make build \
    VERSION=%{version} \
    REVISION=%{release} \
    BRANCH=main \
    BUILD_USER=rpm-builder \
    BUILD_DATE="$(date -u '+%%Y-%%m-%%d_%%H:%%M:%%S')"

%install
# Create directories
install -d %{buildroot}%{_bindir}
install -d %{buildroot}%{_sysconfdir}/%{name}
install -d %{buildroot}%{_unitdir}
install -d %{buildroot}%{_sharedstatedir}/%{name}
install -d %{buildroot}%{_localstatedir}/log/%{name}
install -d %{buildroot}%{_mandir}/man1
install -d %{buildroot}%{_docdir}/%{name}

# Install binary
install -m 0755 bin/%{name} %{buildroot}%{_bindir}/%{name}

# Install configuration file
install -m 0640 packaging/rpm/config.yaml %{buildroot}%{_sysconfdir}/%{name}/config.yaml

# Install systemd service file
install -m 0644 packaging/rpm/%{name}.service %{buildroot}%{_unitdir}/%{name}.service

# Install logrotate configuration
install -d %{buildroot}%{_sysconfdir}/logrotate.d
install -m 0644 packaging/rpm/%{name}.logrotate %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

# Install man page
install -m 0644 packaging/rpm/%{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

# Install documentation
install -m 0644 README.md %{buildroot}%{_docdir}/%{name}/
install -m 0644 CHANGELOG.md %{buildroot}%{_docdir}/%{name}/
install -m 0644 LICENSE %{buildroot}%{_docdir}/%{name}/

# Install example configurations
install -d %{buildroot}%{_docdir}/%{name}/examples
install -m 0644 configs/*.yaml %{buildroot}%{_docdir}/%{name}/examples/

# Install Grafana dashboards
install -d %{buildroot}%{_docdir}/%{name}/dashboards
install -m 0644 dashboards/*.json %{buildroot}%{_docdir}/%{name}/dashboards/

# Install Prometheus alerts
install -d %{buildroot}%{_docdir}/%{name}/alerts
install -m 0644 alerts/*.yml %{buildroot}%{_docdir}/%{name}/alerts/

%pre
# Create system user and group
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || \
    useradd -r -g %{name} -d %{_sharedstatedir}/%{name} -s /sbin/nologin \
    -c "SLURM Exporter service account" %{name}

%post
# Set proper ownership and permissions
chown %{name}:%{name} %{_sharedstatedir}/%{name}
chown %{name}:%{name} %{_localstatedir}/log/%{name}
chown root:%{name} %{_sysconfdir}/%{name}/config.yaml
chmod 640 %{_sysconfdir}/%{name}/config.yaml

# Enable and start systemd service
%systemd_post %{name}.service

%preun
# Stop and disable systemd service
%systemd_preun %{name}.service

%postun
# Clean up systemd service
%systemd_postun_with_restart %{name}.service

# Remove user and group on package removal (not upgrade)
if [ $1 -eq 0 ]; then
    getent passwd %{name} >/dev/null && userdel %{name}
    getent group %{name} >/dev/null && groupdel %{name}
fi

%files
# Binary
%{_bindir}/%{name}

# Configuration
%dir %attr(750, root, %{name}) %{_sysconfdir}/%{name}
%config(noreplace) %attr(640, root, %{name}) %{_sysconfdir}/%{name}/config.yaml

# Systemd service
%{_unitdir}/%{name}.service

# Data directory
%dir %attr(755, %{name}, %{name}) %{_sharedstatedir}/%{name}

# Log directory
%dir %attr(755, %{name}, %{name}) %{_localstatedir}/log/%{name}

# Logrotate configuration
%config(noreplace) %{_sysconfdir}/logrotate.d/%{name}

# Man page
%{_mandir}/man1/%{name}.1*

# Documentation
%doc %{_docdir}/%{name}/README.md
%doc %{_docdir}/%{name}/CHANGELOG.md
%license %{_docdir}/%{name}/LICENSE
%doc %{_docdir}/%{name}/examples/
%doc %{_docdir}/%{name}/dashboards/
%doc %{_docdir}/%{name}/alerts/

%changelog
* Mon Jan 15 2024 SLURM Exporter Team <maintainers@slurm-exporter.io> - 1.0.0-1
- Initial RPM package release
- Comprehensive SLURM metrics collection
- Advanced job analytics engine
- High availability and performance optimizations
- Security hardening and audit compliance
- Kubernetes and container native support

* Thu Dec 14 2023 SLURM Exporter Team <maintainers@slurm-exporter.io> - 0.9.0-1
- Beta release with job analytics
- Performance improvements and caching
- Enhanced security features
- Grafana dashboards and Prometheus alerts

* Tue Nov 21 2023 SLURM Exporter Team <maintainers@slurm-exporter.io> - 0.8.0-1
- Alpha release with basic SLURM metrics
- Docker and Kubernetes support
- Initial documentation and examples