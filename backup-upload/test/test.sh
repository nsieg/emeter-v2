#!/bin/bash
set -e

##########
# ARRANGE
##########

# Cleanup
rm -rf data
rm -rf *.log

echo "Preparing ftp server..."

# Create local files
mkdir data
echo hello > $(pwd)/data/123.txt
echo hello > $(pwd)/data/456.txt

# Start FTP server
docker run --name ftpd -d -p 21:21 -p 30000-30009:30000-30009 -e FTP_USER_HOME=/home/ftpusers/bob -e FTP_USER_NAME=bob -e FTP_USER_PASS=secret stilliard/pure-ftpd

# Create initial files
docker exec ftpd sh -c "echo foo > /home/ftpusers/bob/123.txt"
docker exec ftpd sh -c "echo foo > /home/ftpusers/bob/old.txt"

# List files
docker exec ftpd sh -c "ls -lisah /home/ftpusers/bob" > before.log

##########
# ACT
##########

echo "Sync files to server..."

docker run --rm --name em-backup-lftp -i -e FTP_USERNAME=bob -e FTP_PASSWORD=secret -e FTP_HOST=host.docker.internal -v $(pwd)/data:/data lftp > em-backup-lftp.log 2>&1

##########
# ASSERT
##########

# List files
docker exec ftpd sh -c "ls -lisah /home/ftpusers/bob" > after.log

val123=$(docker exec ftpd sh -c "cat /home/ftpusers/bob/123.txt")
val456=$(docker exec ftpd sh -c "cat /home/ftpusers/bob/123.txt")
valold=$(docker exec ftpd sh -c "cat /home/ftpusers/bob/old.txt")

if [[ $val123 != hello ]]; then
    echo "[FAIL] 123.txt was $val123 but should be foo"
    exit 1
fi
if [[ $val456 != hello ]]; then
    echo "[FAIL] 456.txt was $val456 but should be hello"
    exit 1
fi
if [[ $valold != foo ]]; then
    echo "[FAIL] old.txt was $valold but should be foo"
    exit 1
fi

# Local files should be removed
if test -f $(pwd)/data/123.txt; then
    echo "[FAIL] $(pwd)/data/123.txt should have been deleted"
    exit 1
fi
if test -f $(pwd)/data/456.txt; then
    echo "[FAIL] $(pwd)/data/456.txt should have been deleted"
    exit 1
fi


echo "[SUCCESS] All assertions passed; test successful"

# Get logs
docker logs ftpd > ftpd.log

# Shut down containers
docker rm --force ftpd
