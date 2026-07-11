#!/bin/sh

# AdGuard Home Installation Script
# For the quydang04/AdGuardHome fork.

# Exit the script if a pipeline fails (-e), prevent accidental filename
# expansion (-f), and consider undefined variables as errors (-u).
set -e -f -u

repo_owner='quydang04'
repo_name='AdGuardHome'
repo_branch='master'
readonly repo_owner repo_name repo_branch

log() {
	if [ "$verbose" -gt '0' ]; then
		echo "$1" 1>&2
	fi
}

error_exit() {
	echo "$1" 1>&2

	exit 1
}

usage() {
	echo 'install.sh: usage: [-c channel] [-C cpu_type] [-h] [-O os] [-o output_dir]' \
		'[-r|-R] [-u|-U] [-v|-V]' 1>&2

	exit 2
}

maybe_sudo() {
	if [ "$use_sudo" -eq 0 ]; then
		"$@"
	else
		"$sudo_cmd" "$@"
	fi
}

is_command() {
	command -v "$1" >/dev/null 2>&1
}

is_little_endian() {
	is_little_endian_result="$(
		printf 'I' \
			| hexdump -o \
			| awk '{ print substr($2, 6, 1); exit; }'
	)"
	readonly is_little_endian_result

	[ "$is_little_endian_result" -eq '1' ]
}

check_required() {
	required_darwin="unzip"
	required_unix="tar"
	readonly required_darwin required_unix

	case "$os" in
	'freebsd' | 'linux' | 'openbsd')
		required="$required_unix"
		;;
	'darwin')
		required="$required_darwin"
		;;
	*)
		error_exit "unsupported operating system: '$os'"
		;;
	esac
	readonly required

	for cmd in $required; do
		log "checking $cmd"
		if ! is_command "$cmd"; then
			log "the full list of required software: [$required]"

			error_exit "$cmd is required to install AdGuard Home via this script"
		fi
	done
}

check_out_dir() {
	if [ "$out_dir" = '' ]; then
		error_exit 'output directory should be presented'
	fi

	if ! [ -d "$out_dir" ]; then
		log "$out_dir directory will be created"
	fi
}

parse_opts() {
	while getopts "C:c:hO:o:rRuUvV" opt "$@"; do
		case "$opt" in
		C)
			cpu="$OPTARG"
			;;
		c)
			channel="$OPTARG"
			;;
		h)
			usage
			;;
		O)
			os="$OPTARG"
			;;
		o)
			out_dir="$OPTARG"
			;;
		R)
			reinstall='0'
			;;
		U)
			uninstall='0'
			;;
		r)
			reinstall='1'
			;;
		u)
			uninstall='1'
			;;
		V)
			verbose='0'
			;;
		v)
			verbose='1'
			;;
		*)
			log "bad option $OPTARG"

			usage
			;;
		esac
	done

	if [ "$uninstall" -eq '1' ] && [ "$reinstall" -eq '1' ]; then
		error_exit 'the -r and -u options are mutually exclusive'
	fi
}

set_channel() {
	case "$channel" in
	'release' | 'development')
		;;
	*)
		error_exit "invalid channel '$channel'
supported values are 'release' and 'development'"
		;;
	esac

	log "channel: $channel"
}

set_os() {
	if [ "$os" = '' ]; then
		os="$(uname -s)"
		case "$os" in
		'Darwin')
			os='darwin'
			;;
		'FreeBSD')
			os='freebsd'
			;;
		'Linux')
			os='linux'
			;;
		'OpenBSD')
			os='openbsd'
			;;
		*)
			error_exit "unsupported operating system: '$os'"
			;;
		esac
	fi

	case "$os" in
	'darwin' | 'freebsd' | 'linux' | 'openbsd')
		;;
	*)
		error_exit "unsupported operating system: '$os'"
		;;
	esac

	log "operating system: $os"
}

set_cpu() {
	if [ "$cpu" = '' ]; then
		cpu="$(uname -m)"
		case "$cpu" in
		'x86_64' | 'x86-64' | 'x64' | 'amd64')
			cpu='amd64'
			;;
		'i386' | 'i486' | 'i686' | 'i786' | 'x86')
			cpu='386'
			;;
		'armv5l')
			cpu='armv5'
			;;
		'armv6l')
			cpu='armv6'
			;;
		'armv7l' | 'armv8l')
			cpu='armv7'
			;;
		'aarch64' | 'arm64')
			cpu='arm64'
			;;
		'mips' | 'mips64')
			if is_little_endian; then
				cpu="${cpu}le"
			fi

			cpu="${cpu}_softfloat"
			;;
		'riscv64')
			cpu='riscv64'
			;;
		*)
			error_exit "unsupported cpu type: $cpu"
			;;
		esac
	fi

	case "$cpu" in
	'amd64' | '386' | 'armv5' | 'armv6' | 'armv7' | 'arm64' | 'riscv64')
		;;
	'mips64le_softfloat' | 'mips64_softfloat' | 'mipsle_softfloat' | 'mips_softfloat')
		;;
	*)
		error_exit "unsupported cpu type: $cpu"
		;;
	esac

	log "cpu type: $cpu"
}

fix_darwin() {
	if [ "$os" != 'darwin' ]; then
		return 0
	fi

	pkg_ext='zip'
	out_dir='/Applications'
}

fix_freebsd() {
	if [ "$os" != 'freebsd' ]; then
		return 0
	fi

	rcd='/usr/local/etc/rc.d'
	readonly rcd

	if ! [ -d "$rcd" ]; then
		mkdir "$rcd"
	fi
}

download_curl() {
	curl_output="${2:-}"
	if [ "$curl_output" = '' ]; then
		curl -L -S -s "$1"
	else
		curl -L -S -o "$curl_output" -s "$1"
	fi
}

download_wget() {
	wget_output="${2:--}"

	wget --no-verbose -O "$wget_output" "$1"
}

download_fetch() {
	fetch_output="${2:-}"
	if [ "$fetch_output" = '' ]; then
		fetch -o '-' "$1"
	else
		fetch -o "$fetch_output" "$1"
	fi
}

set_download_func() {
	if is_command 'curl'; then
		return 0
	elif is_command 'wget'; then
		download_func='download_wget'
	elif is_command 'fetch'; then
		download_func='download_fetch'
	else
		error_exit "either curl or wget is required to install AdGuard Home via this script"
	fi
}

set_sudo_cmd() {
	case "$os" in
	'openbsd')
		sudo_cmd='doas'
		;;
	'darwin' | 'freebsd' | 'linux')
		;;
	*)
		error_exit "unsupported operating system: '$os'"
		;;
	esac
}

configure() {
	set_channel
	set_os
	set_cpu
	fix_darwin
	set_download_func
	set_sudo_cmd
	check_out_dir

	pkg_name="AdGuardHome_${os}_${cpu}.${pkg_ext}"
	release_base_url="https://github.com/${repo_owner}/${repo_name}/releases"
	url="${release_base_url}/latest/download/${pkg_name}"
	agh_dir="${out_dir}/AdGuardHome"
	readonly pkg_name url agh_dir release_base_url

	log "AdGuard Home will be installed from $url"
	log "AdGuard Home will be installed into $agh_dir"
}

is_root() {
	user_id="$(id -u)"
	if [ "$user_id" -eq '0' ]; then
		log 'script is executed with root privileges'

		return 0
	fi

	if is_command "$sudo_cmd"; then
		log 'note that AdGuard Home requires root privileges to install using this script'

		return 1
	fi

	error_exit 'root privileges are required to install AdGuard Home using this script
please, restart it with root privileges'
}

rerun_with_root() {
	script_url="https://raw.githubusercontent.com/${repo_owner}/${repo_name}/${repo_branch}/scripts/install.sh"
	readonly script_url

	r='-R'
	if [ "$reinstall" -eq '1' ]; then
		r='-r'
	fi

	u='-U'
	if [ "$uninstall" -eq '1' ]; then
		u='-u'
	fi

	v='-V'
	if [ "$verbose" -eq '1' ]; then
		v='-v'
	fi

	readonly r u v

	log 'restarting with root privileges'

	{ "$download_func" "$script_url" || echo 'exit 1'; } \
		| $sudo_cmd sh -s -- -c "$channel" -C "$cpu" -O "$os" -o "$out_dir" "$r" "$u" "$v"

	exit 0
}

download() {
	log "downloading package from $url to $pkg_name"

	if ! "$download_func" "$url" "$pkg_name"; then
		error_exit "cannot download the package from $url into $pkg_name"
	fi

	log "successfully downloaded $pkg_name"
}

unpack() {
	log "unpacking package from $pkg_name into $out_dir"

	# shellcheck disable=SC2174
	if ! mkdir -m 0700 -p "$out_dir"; then
		error_exit "cannot create directory $out_dir"
	fi

	case "$pkg_ext" in
	'zip')
		unzip "$pkg_name" -d "$out_dir"
		;;
	'tar.gz')
		tar -C "$out_dir" -f "$pkg_name" -x -z
		;;
	*)
		error_exit "unexpected package extension: '$pkg_ext'"
		;;
	esac

	unpacked_contents="$(
		echo
		ls -l -A "$agh_dir"
	)"
	log "successfully unpacked, contents: $unpacked_contents"

	rm "$pkg_name"
}

handle_existing() {
	if ! [ -d "$agh_dir" ]; then
		log 'no need to uninstall'

		if [ "$uninstall" -eq '1' ]; then
			exit 0
		fi

		return 0
	fi

	existing_adguard_home="$(ls -1 -A "$agh_dir")"
	if [ "$existing_adguard_home" != '' ]; then
		log 'the existing AdGuard Home installation is detected'

		if [ "$reinstall" -ne '1' ] && [ "$uninstall" -ne '1' ]; then
			error_exit \
				"to reinstall/uninstall the AdGuard Home using this script specify one of the '-r' or '-u' flags"
		fi

		if (cd "$agh_dir" && ! ./AdGuardHome -s stop || ! ./AdGuardHome -s uninstall); then
			log "cannot uninstall AdGuard Home from $agh_dir"
		fi

		rm -r "$agh_dir"

		log 'AdGuard Home was successfully uninstalled'
	fi

	if [ "$uninstall" -eq '1' ]; then
		exit 0
	fi
}

install_service() {
	use_sudo='0'
	if [ "$os" = 'freebsd' ]; then
		use_sudo='1'
	fi

	if (cd "$agh_dir" && maybe_sudo ./AdGuardHome -s install); then
		return 0
	fi

	log "installation failed, removing $agh_dir"

	rm -r "$agh_dir"

	if [ "$cpu" = 'armv7' ]; then
		cpu='armv5'
		reinstall='1'

		log "trying to use $cpu cpu"

		rerun_with_root
	fi

	error_exit 'cannot install AdGuardHome as a service'
}

# Entrypoint

channel='release'
reinstall='0'
uninstall='0'
verbose='0'
cpu=''
os=''
out_dir='/opt'
pkg_ext='tar.gz'
download_func='download_curl'
sudo_cmd='sudo'

parse_opts "$@"

echo 'starting AdGuard Home installation script'

configure
check_required

if ! is_root; then
	rerun_with_root
fi
fix_freebsd

handle_existing

download
unpack

install_service

printf '%s\n' \
	'AdGuard Home is now installed and running' \
	'you can control the service status with the following commands:' \
	"$sudo_cmd ${agh_dir}/AdGuardHome -s start|stop|restart|status|install|uninstall"
