class zabbix::params {
  $basedir = '/etc/zabbix'
  $config = "${basedir}/zabbix_agentd.conf"
  $config_dir = "${basedir}/zabbix_agentd.d"
  $scripts_dir = "${basedir}/scripts"
  $pidfile = '/var/run/zabbix/zabbix_agentd.pid'
  $logfile = '/var/log/zabbix/zabbix_agentd.log'
  $logfile_size = 0
  $metadata = "kernel=${::kernel};osfamily=${::osfamily};os=${::operatingsystem};osversion=${::operatingsystemrelease};"
}
