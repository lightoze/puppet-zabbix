class zabbix::jolokia (
  $cache_dir = "/var/cache/zabbix-jolokia"
) {
  include zabbix::agent

  $basedir = dirname($zabbix::agent::config_dir)

  $read_script = "${zabbix::agent::scripts_dir}/jolokia-read"
  $discovery_script = "${zabbix::agent::scripts_dir}/jolokia-discovery"
  file { $read_script:
    source => "puppet:///modules/zabbix/jolokia-read.${::os['architecture']}",
    mode   => '0755',
  }
  file { $discovery_script:
    source => "puppet:///modules/zabbix/jolokia-discovery.${::os['architecture']}",
    mode   => '0755',
  }

  file { $cache_dir:
    ensure => directory,
    owner  => 'zabbix',
    group  => 'zabbix',
    mode   => '0750',
  }

  file { "${zabbix::agent::config_dir}/jolokia.conf":
    content => template('zabbix/jolokia.conf.erb'),
  } ~> Service['zabbix-agent']
}
