class zabbix::postgresql (
  $binary_path = '',
  $connect_options = '-d postgres',
) {
  include zabbix::agent

  file { "${zabbix::agent::config_dir}/postgresql.conf":
    content => template('zabbix/postgresql.conf.erb'),
  } ~> Service['zabbix-agent']

  if defined(Class['postgresql::server']) {
    postgresql::server::role { 'zabbix':
      login => true,
    }
    postgresql::server::pg_hba_rule { 'zabbix':
      type        => 'local',
      database    => 'postgres',
      user        => 'zabbix',
      auth_method => 'ident',
    }
  }
}
