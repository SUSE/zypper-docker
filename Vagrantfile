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

    # Install the latest zypper and docker.
    zypper ar http://download.opensuse.org/repositories/zypp:/Head/openSUSE_13.2/ zypp:Head
    zypper --gpg-auto-import-keys ref && zypper -n remove yast2-pkg-bindings libyui-ncurses-pkg6
    zypper --gpg-auto-import-keys ref && zypper -n install --from zypp:Head zypper libzypp
    zypper --gpg-auto-import-keys ref && zypper -n install docker

    # Since Go 1.4 is not in openSUSE 13.2 yet, let's use the one as provided
    # by the Go project.
    wget -q https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz -O go.tar.gz
    tar xzf go.tar.gz && rm go.tar.gz
    echo 'export GOROOT=/home/vagrant/go' >> /home/vagrant/.bashrc
    echo 'export GOPATH=/home/vagrant/gopath' >> /home/vagrant/.bashrc
    echo 'export PATH=\$GOROOT/bin:\$GOPATH/bin:\$PATH' >> /home/vagrant/.bashrc
    source /home/vagrant/.bashrc

    # Install zypper-docker
    cd $GOPATH/src/github.com/SUSE/zypper-docker
    go get github.com/tools/godep
    godep restore
    go install

    # Finally enable & start docker
    /usr/sbin/usermod -G docker vagrant
    systemctl enable docker
    systemctl start docker
EOS
  end

  config.vm.synced_folder '.', '/vagrant', disabled: true
  config.vm.synced_folder '.', '/home/vagrant/gopath/src/github.com/SUSE/zypper-docker'
end

