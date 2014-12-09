class zabbix::postgresql {
    include zabbix::agent

    file { "${zabbix::agent::config_dir}/postgresql.conf":
        source => 'puppet:///modules/zabbix/postgresql.conf',
    } ~> Service['zabbix-agent']

    if defined(Class['postgresql::server']) {
        postgresql::server::role { 'zabbix':
            login => true,
        }
        postgresql::server::pg_hba_rule { 'zabbix':
            type => 'local',
            database => 'postgres',
            user => 'zabbix',
            auth_method => 'ident',
        }
    }
}
