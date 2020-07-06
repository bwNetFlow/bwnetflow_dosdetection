# bwnetflow_dosdetection

The bwNetFlow Dos Detection is a sample implementation of a volume-based (D)Dos Detection application using the bwNetFlow exported flows.
The sample implementation consists of a docker-compose file setting up several docker images/container.

It consists of the following images serving the respective function:grafana
* bandwidth calculator **bw** (./container/bw): calculates the current bandwidth (seperated into up- and downlink) in the specified network. It sets a gauge to the current current bandwidth (IMPORTANT: it is not necessary for the DoS detection; it just servers as convenient overview in the grafana dashboard.)
* threshold calculator **thresholds** (./container/thresholds): calculates the thresholds used for the DoS Detection over the specified period of time. Uses the specified threshold multiplicator as *buffer*. Writes the thresholds to a file that is read by the DoS detector; also sets a gauge with time labels to the current threshold.
* DoS detector **detection** (./container/detection): calculates the current peered bandwidth in the network and compares it with the respective threshold. If the current peered bandwidth exceeds the thresolds it sets a prometheus gauge to the current peered bandwidth; else the gauge is zero.
* prometheus server **prometheus** (./container/prometheus): collects all gauges written by the before mentioned container and provides the data to Grafana.
* Grafana **grafana** (./container/grafana): collects the data provided by the Prometheus container; sets up the Grafana dashboard over HTTPS.

# Settings.ini
The mandatory settings that must be defined before the making.  
The respective file is located in *./container/general_conf*.  
The following parameters can/must be defined:  
```
topic=<kafka topic> (MUST; no default)
user=<kafka user> (MUST; no default)
pwd=<user's password> (MUST; no default)
bw_grp_id=<kafka ID for the bw container consumer> (MUST; no default)
threshold_grp_id=<kafka ID for the threshold container consumer> (MUST; no default)
detection_grp_id=<kafka ID for the detection container consumer> (MUST; no default)
brokers=<kafka server IP:PORT> (MUST; no default)
timezone=<the user's timezone according to the IANA zime zone database> (MUST; no default)
training_time=<time the threshold container consumes flows to calculate the DoS threshold in seconds> (OPTIONAL; default: 86400)
threshold_multiplier=<the multiplier that is applied to the max bandwidth calculated during the training time to build a buffer> (OPTIONAL; default: 2)
```
# File Permissions
The respective container user using the directories in located in ./data must be provided the necessary permissions. 
* The *./data/threholds* directory is used by the container *threholds*, which needs write and execute permissions. The user running the *detection* container needs read and execute permissions.
* The *./data/grafana* directory is used by the container *grafana*. The user running this container needs read, write and execute permissions.
* The *./data/prometheus* directory is used by the container *pserver*. The user running this container needs read, write and execute permissions.

# Grafana SSL Credentials
Do not forget to create an priv key and ssl cert

# Installation
You can create all necessary images by just typing 
```
make all
```
NOTE: It needs a valid *./container/general_conf/settings.ini* file for successfully making the images.

# Starting
After making all images the sample implementation can be started by 
```
sudo docker-compose up -d
```
