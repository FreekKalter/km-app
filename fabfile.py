from fabric.api import *
import os

env.use_ssh_config = True
env.ssh_config_path = '/var/lib/jenkins/.ssh/config'
env.hosts.extend(['fkalter@km-app.dyndns.org'])


def deploy():
    prepareDeploy()
    remoteDeploy()

def prepareDeploy():
    buildNumber = os.environ['BUILD_NUMBER']
    local("make prepare-production")
    local("docker build -t freekkalter/km:{} .".format(buildNumber))
    local("docker push freekkalter/km")

def remoteDeploy():
    buildNumber = os.environ['BUILD_NUMBER']
    run("docker pull freekkalter/km")
    runBuildNr(buildNumber)

def rollback():
    # find latest buildnumber on remote, default = the build before the last one
    buildNumber = int(prompt('Rever to buildnumber: ', validate=int, default=getLatestBuildNr()-1))
    if buildNumber < 0:
        buildNumber = bn+buildNumber
    runBuildNr(buildNumber)

def runBuildNr(buildNumber):
    cidfile = '/home/fkalter/.km.cidfile'
    run("docker kill `cat {}`".format(cidfile))
    run("rm {}".format(cidfile))
    run("docker rm km_production")
    run("docker run -name km_production -cidfile={}\
                    -v /home/fkalter/km/postgresdata:/data:rw\
                    -v /home/fkalter/km/log:/log\
                    -d -p 4001:4001 \
                    freekkalter/km:{} /usr/bin/supervisord".format(cidfile, buildNumber))

def getSqlDump(directory):
    run('docker run -v /home/fkalter/backup:/backup:rw -link km_production:main freekkalter/km:{} /backup.sh'.format(getLatestBuildNr()))
    get('/home/fkalter/backup/backup.sql', directory)

def pullProductionData():
    local('mkdir -p backup')
    getSqlDump('./backup')

    # import into local running container
    local('docker run -v /home/fkalter/github/km/backup:/backup:rw -link km_production:main freekkalter/km:deploy /restore.sh')

# call backup from cronjob/jenkins
def backup():
    getSqlDump('.')
    # tar file and move to folder with format backup-{date}.sql
    local("tar -czf ~/km-backup/backup_`date +%d-%m-%Y.tar.gz` ./backup.sql")
    local("rm ./backup.sql")

def getLatestBuildNr():
    return int(run("docker images | awk '{ if(match($2, /^[0-9]+$/)) print $2}' | sort | tail -n1"))
