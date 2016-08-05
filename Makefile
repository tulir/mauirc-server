default:
	go install

build:
	go build -o mauircd

package-prep: build
	mkdir -p package/usr/bin/
	mkdir -p package/etc/mauircd/
	mkdir -p package/var/log/mauircd/
	cp mauircd package/usr/bin/
	cp example-config.json package/etc/mauircd/config.json

package: package-prep
	dpkg-deb --build package mauircd.deb > /dev/null

clean:
	rm -rf mauircd mauircd.deb package/usr/bin package/var/log/mauircd package/etc/mauircd
