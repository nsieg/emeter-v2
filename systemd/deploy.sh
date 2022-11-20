#!/bin/bash
read -sp 'Password: ' passvar

sshtarget=nils@192.168.168.26
scpcmd="SSHPASS=$passvar sshpass -e scp /home/nils/projects/emeter"
sshcmd="SSHPASS=$passvar sshpass -e ssh $sshtarget 'echo $passvar | sudo -S"

# vzlogger initial setup
eval "$scpcmd/meter-ingest/lib* $sshtarget/usr/local/lib/"
eval "$scpcmd/systemd/vzlogger $sshtarget/opt/energymeter"
# backup-export
eval "$scpcmd/backup/backup-export/em-backup-export $sshtarget/opt/energymeter"
eval "$scpcmd/backup/backup-export/*.tmpl $sshtarget/opt/energymeter"
# backup-upload
eval "$scpcmd/backup/backup-upload/upload.sh $sshtarget/opt/energymeter"
# em-shelly
eval "$scpcmd/shelly-ingest/em-shelly $sshtarget/opt/energymeter"
# em-report
eval "$scpcmd/report/em-report $sshtarget/opt/energymeter"
# vzlogger-conf
eval "$scpcmd/meter-ingest/vzlogger-man.conf $sshtarget/opt/energymeter/vzlogger.conf"
# emeter.env
eval "$scpcmd/systemd/emeter.env $sshtarget/opt/energymeter/emeter.env"
# service + timer
eval "$scpcmd/systemd/*.service $sshtarget/etc/systemd/system"
eval "$scpcmd/systemd/*.timer $sshtarget/etc/systemd/system"

eval "$sshcmd service enable em-backup-export'"
eval "$sshcmd service enable em-backup-upload'"
eval "$sshcmd service enable em-shelly'"
eval "$sshcmd service enable em-report'"
eval "$sshcmd service enable em-vzlogger'"

eval "$sshcmd service restart em-backup-export'"
eval "$sshcmd service restart em-backup-upload'"
eval "$sshcmd service restart em-shelly'"
eval "$sshcmd service restart em-report'"
eval "$sshcmd service restart em-vzlogger'"

