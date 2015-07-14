# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure('2') do |config|
  config.vm.provider 'virtualbox' do |vb|
    # Use VBoxManage to customize the VM. For example to change memory:
    vb.customize ['modifyvm', :id, '--memory', '1024']
    # Useful when something bad happens
    # vb.gui = true
  end

  config.vm.define :test_docker do |node|
    node.vm.box = 'flavio/opensuse13-2'
    node.vm.box_check_update = true
    node.vm.provision 'shell', inline: <<EOS
    echo "Hello world"
    zypper ar http://download.opensuse.org/repositories/zypp:/Head/openSUSE_13.2/ zypp:Head
    zypper --gpg-auto-import-keys ref && zypper -n install --from zypp:Head zypper libzypp
EOS
  end
end

