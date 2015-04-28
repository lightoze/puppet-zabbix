class zabbix::agent::params {
    $config = '/etc/zabbix/zabbix_agentd.conf'
    $config_dir = '/etc/zabbix/zabbix_agentd.d'
    $pidfile = '/var/run/zabbix/zabbix_agentd.pid'
    $logfile = '/var/log/zabbix/zabbix_agentd.log'
    $logfile_size = 0
    $metadata = "kernel=${::kernel};osfamily=${::osfamily};os=${::operatingsystem};osversion=${::operatingsystemrelease};"
}

class zabbix::agent (
    $version = '2.2',
    $config = $::zabbix::agent::params::config,
    $config_dir = $zabbix::agent::params::config_dir,
    $pidfile = $::zabbix::agent::params::pidfile,
    $logfile = $::zabbix::agent::params::logfile,
    $logfile_size = $::zabbix::agent::params::logfile_size,
    $metadata = $::zabbix::agent::params::metadata,
    $remote_commands = disabled,
    $source_ip = undef,
    $listen_port = 10050,
    $listen_ip = undef,
    $server,
    $server_active = undef,
    $agents = 3,
    $timeout = 5,
    $unsafe_parameters = false,
) inherits zabbix::agent::params {
    validate_re($remote_commands, '^(disabled|enabled|log)$')

    case $::osfamily {
        default: { fail("Unsupported platform ${::osfamily}") }
        'Debian': {
            $os = downcase($::operatingsystem)
            apt::source { 'zabbix':
                location => "http://repo.zabbix.com/zabbix/${version}/${os}",
                release => "${::lsbdistcodename}",
                repos => 'main',
                key => '79EA5ED4',
                key_server => 'pgp.mit.edu',
            }
            apt::hold { 'zabbix-agent':
                version => "${version}.*",
            }
            Class['apt::update'] -> Package['zabbix-agent']
        }
        'RedHat': {
            yumrepo { 'zabbix':
                descr => "Zabbix",
                baseurl => "http://repo.zabbix.com/zabbix/${version}/rhel/${::operatingsystemmajrelease}/\$basearch",
                gpgcheck => true,
                gpgkey => "http://repo.zabbix.com/RPM-GPG-KEY-ZABBIX",
            }
            Yumrepo['zabbix'] -> Package['zabbix-agent']
        }
    }

    package { 'zabbix-agent':
        ensure => latest,
    }
    ->
    file { $config:
        ensure => file,
        content => template('zabbix/zabbix_agentd.conf.erb'),
    }
    ~>
    service { 'zabbix-agent':
        ensure => running,
        enable => true,
    }

    file { $config_dir:
        ensure => directory,
        recurse => true,
        purge => true,
        require => Package['zabbix-agent'],
    }
    file { "$config_dir/sysinfo.conf":
        source => 'puppet:///modules/zabbix/sysinfo.conf',
    } ~> Service['zabbix-agent']

    if defined(Class['postgresql::server']) {
        include zabbix::postgresql
    }
}
