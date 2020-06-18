all: copy build clean

copy:
	cp container/general_conf/settings.ini container/bw
	cp container/general_conf/settings.ini container/thresholds
	cp container/general_conf/settings.ini container/detection

build:
	cd container/bw && $(MAKE) all
	cd container/thresholds && $(MAKE) all
	cd container/detection && $(MAKE) all
	cd container/prometheus && $(MAKE) all

clean:
	rm container/bw/settings.ini
	rm container/thresholds/settings.ini
	rm container/detection/settings.ini
