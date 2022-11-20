#!/bin/bash
set -e

lftp -c "set ftp:ssl-allow no; open -u $FTP_USERNAME,$FTP_PASSWORD $FTP_HOST; mirror -R $DATA_DIR / --parallel=10"

rm $DATA_DIR/*