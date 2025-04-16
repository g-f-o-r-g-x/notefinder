Name:           notefinder
Version:        0.1
Release:        1%{?dist}
Summary:        A note-taking application

License:        BSD-3-Clause
URL:            https://github.com/g-f-o-r-g-x/notefinder
Source0:        %{name}-%{version}.tar.gz

BuildRequires: go
BuildRequires: gcc
BuildRequires: autoconf
BuildRequires: automake
BuildRequires: make

%if 0%{?suse_version}
BuildRequires: perl
%else
BuildRequires: perl-devel
BuildRequires: perl-ExtUtils-Embed
%endif

%if 0%{?suse_version}
BuildRequires: libXcursor-devel
BuildRequires: libXrandr-devel
BuildRequires: Mesa-libGL-devel
BuildRequires: libXi-devel
BuildRequires: libXinerama-devel
BuildRequires: libXxf86vm-devel
BuildRequires: libxkbcommon-devel
BuildRequires: wayland-devel
%else
BuildRequires: libXcursor-devel
BuildRequires: libXrandr-devel
BuildRequires: mesa-libGL-devel
BuildRequires: libXi-devel
BuildRequires: libXinerama-devel
BuildRequires: libXxf86vm-devel
BuildRequires: libxkbcommon-devel
BuildRequires: wayland-devel
%endif

BuildRequires: desktop-file-utils

Requires:       hicolor-icon-theme
Requires:       perl

%description
A note-taking/personal information management application that simplifies
keeping, organizing and searching various pieces of data: notes, bookmarks,
documents, tasks from different sources in one place.
Extensible with Perl scripts

%prep
%autosetup

%build
autoreconf -i
./configure --prefix=/usr
make

%install
make install DESTDIR=%{buildroot}

%check
desktop-file-validate %{buildroot}%{_datadir}/applications/notefinder.desktop

%files
%license LICENSE.md
%doc README.md
%{_bindir}/notefinder
%{perl_vendorlib}/Notefinder.pm
%{_datadir}/applications/notefinder.desktop
%{_datadir}/icons/hicolor/64x64/apps/notefinder.png

%changelog
* Tue Apr 15 2025 Sergey S. <janis19011943@gmail.com.com> - 0.1-1
- Initial package
