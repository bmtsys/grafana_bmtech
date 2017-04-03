#!/bin/bash

# sqlite3
# cookie_string=grafana_sess=c0ac6143272c6983;

# mysql
# cookie_string=grafana_sess=0d47d5d9d0d3e42b;


cookie_string='grafana_sess=0d47d5d9d0d3e42b=&uid=1;'

ab -n 20000 -c 20 -C $cookie_string http://localhost/grafana/api/datasources/proxy/1/metrics/find/?query=apps.*

# Sqlite3
# Concurrency Level:      20
# Time taken for tests:   18.641 seconds
# Complete requests:      20000
# Failed requests:        0
# Total transferred:      14080000 bytes
# HTML transferred:       6400000 bytes
# Requests per second:    1072.93 [#/sec] (mean)
# Time per request:       18.641 [ms] (mean)
# Time per request:       0.932 [ms] (mean, across all concurrent requests)
# Transfer rate:          737.64 [Kbytes/sec] received

# mysql
# Concurrency Level:      20
# Time taken for tests:   24.054 seconds
# Complete requests:      20000
# Failed requests:        0
# Total transferred:      14080000 bytes
# HTML transferred:       6400000 bytes
# Requests per second:    831.48 [#/sec] (mean)
# Time per request:       24.054 [ms] (mean)
# Time per request:       1.203 [ms] (mean, across all concurrent requests)
# Transfer rate:          571.64 [Kbytes/sec] received
#

