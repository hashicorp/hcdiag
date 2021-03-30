cd ..

# test local
go run .
cat ./results.json

# build linux binary
GOOS=linux GOARCH=amd64 go build -o host-diagnostics_linux

# spin up ubuntu vm, run binary, check output, tear down
cd ./tests/vagrant_ubuntu && vagrant up

vagrant ssh -c "/home/vagrant/host-diagnostics_linux"
vagrant ssh -c "cat /home/vagrant/results.json"
vagrant destroy -f

# spin up rhel vm, run binary, check output, tear down
cd ../vagrant_rhel && vagrant up
vagrant ssh -c "/home/vagrant/host-diagnostics_linux"
vagrant ssh -c "cat /home/vagrant/results.json"
vagrant destroy -f

cd ..
