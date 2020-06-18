# bwnetflow_dosdetection

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
# Grafana SSL Credentials

Do not forget to create an priv key and ssl cert
