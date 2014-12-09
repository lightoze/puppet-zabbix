class zabbix::jmx {
    include zabbix::agent

    $basedir = dirname($zabbix::agent::config_dir)
    $discovery_script = "${basedir}/jmx-discovery.py"

    file { $discovery_script:
        source => 'puppet:///modules/zabbix/jmx-discovery.py'
    }

    file { "${zabbix::agent::config_dir}/jmx.conf":
        content => template('zabbix/jmx.conf.erb'),
    } ~> Service['zabbix-agent']
}
