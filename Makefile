default:
	go install

build:
	go build -o mauirc-server

package-prep: build
	mkdir -p package/usr/bin/
	mkdir -p package/etc/mauirc/
	mkdir -p package/var/log/mauirc/
	cp mauirc package/usr/bin/
	cp example-config.json package/etc/mauirc/config.json

package: package-prep
	dpkg-deb --build package mauirc-server.deb > /dev/null

clean:
	rm -rf mauirc-server mauirc-server.deb package/usr/bin package/var/log/mauirc package/etc/mauirc
