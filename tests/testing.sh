echo "[$(date +"%FT%T")] 'cd ..'"
cd ..

# test local
echo "[$(date +"%FT%T")] 'go run .'"
go run .

# check local output
echo "[$(date +"%FT%T")] 'cat ./results.json'"
cat ./results.json

# build linux binary
echo "[$(date +"%FT%T")] 'GOOS=linux GOARCH=amd64 go build -o host-diagnostics_linux'"
GOOS=linux GOARCH=amd64 go build -o host-diagnostics_linux

# spin up ubuntu vm, run binary, check output, tear down
echo "[$(date +"%FT%T")] 'cd ./vagrant_ubuntu && vagrant up'"
cd ./tests/vagrant_ubuntu && vagrant up

echo "[$(date +"%FT%T")] 'vagrant ssh -c "/home/vagrant/host-diagnostics_linux"'"
vagrant ssh -c "/home/vagrant/host-diagnostics_linux"

echo "[$(date +"%FT%T")] 'vagrant ssh -c "cat /home/vagrant/results.json"'"
vagrant ssh -c "cat /home/vagrant/results.json"

echo "[$(date +"%FT%T")] 'vagrant destroy -f'"
vagrant destroy -f

echo "[$(date +"%FT%T")] 'cd ../vagrant_rhel && vagrant up'"
cd ../vagrant_rhel && vagrant up

echo "[$(date +"%FT%T")] 'vagrant ssh -c "/home/vagrant/host-diagnostics_linux"'"
vagrant ssh -c "/home/vagrant/host-diagnostics_linux"

echo "[$(date +"%FT%T")] 'vagrant ssh -c "cat /home/vagrant/results.json"'"
vagrant ssh -c "cat /home/vagrant/results.json"

echo "[$(date +"%FT%T")] 'vagrant destroy -f'"
vagrant destroy -f

echo "[$(date +"%FT%T")] 'cd ..'"
cd ..
