require 'spec_helper'
describe 'zabbix::agent' do

  context 'with defaults for all parameters' do
    it { should contain_class('zabbix::agent') }
  end
end
