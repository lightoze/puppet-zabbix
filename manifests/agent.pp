class zabbix::agent::params {
  $basedir = '/etc/zabbix'
  $config = "${basedir}/zabbix_agentd.conf"
  $config_dir = "${basedir}/zabbix_agentd.d"
  $scripts_dir = "${basedir}/scripts"
  $pidfile = '/var/run/zabbix/zabbix_agentd.pid'
  $logfile = '/var/log/zabbix/zabbix_agentd.log'
  $logfile_size = 0
  $metadata = "kernel=${::kernel};osfamily=${::osfamily};os=${::operatingsystem};osversion=${::operatingsystemrelease};"
}

class zabbix::agent (
  $version = '2.2',
  $config = $::zabbix::agent::params::config,
  $config_dir = $zabbix::agent::params::config_dir,
  $scripts_dir = $zabbix::agent::params::scripts_dir,
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
  $allow_root = false,
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
        location   => "http://repo.zabbix.com/zabbix/${version}/${os}",
        release    => "${::lsbdistcodename}",
        repos      => 'main',
        key        => 'FBABD5FB20255ECAB22EE194D13D58E479EA5ED4',
        key_server => 'pgp.mit.edu',
      }
      apt::pin { 'zabbix-agent':
        packages => 'zabbix-agent',
        version  => "${version}.*",
        priority => 1001,
      }
      Class['apt::update'] -> Package['zabbix-agent']
    }
    'RedHat': {
      yumrepo { 'zabbix':
        descr    => "Zabbix",
        baseurl  => "http://repo.zabbix.com/zabbix/${version}/rhel/${::operatingsystemmajrelease}/\$basearch",
        gpgcheck => true,
        gpgkey   => "http://repo.zabbix.com/RPM-GPG-KEY-ZABBIX",
      }
      Yumrepo['zabbix'] -> Package['zabbix-agent']
    }
  }

  package { 'zabbix-agent':
    ensure => latest,
  }
  ->
  file { $config:
    ensure  => file,
    content => template('zabbix/zabbix_agentd.conf.erb'),
  }
  ~>
  service { 'zabbix-agent':
    ensure => running,
    enable => true,
  }

  user { 'zabbix':
    groups  => ['puppet'],
    require => Package['zabbix-agent'],
  }

  file { $config_dir:
    ensure  => directory,
    recurse => true,
    purge   => true,
    require => Package['zabbix-agent'],
  }
  ->
  file { $scripts_dir:
    ensure  => directory,
    recurse => true,
    purge   => true,
  }

  file { "${scripts_dir}/read_yaml.sh":
    ensure => file,
    source => 'puppet:///modules/zabbix/read_yaml.sh',
    mode   => '755',
  }
  file { "${config_dir}/sysinfo.conf":
    content => template('zabbix/sysinfo.conf.erb'),
  } ~> Service['zabbix-agent']

  ensure_resource('package', 'pciutils', { 'ensure' => 'present' })

  if defined(Class['postgresql::server']) {
    include zabbix::postgresql
  }
}
