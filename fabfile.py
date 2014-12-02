from fabric.api import *
import re
import time

env.ssh_config_path = '~/.ssh/config'
env.use_ssh_config = True
env.hosts.extend(['core'])


def deploy_nginx():
    with lcd('nginx'):
        local('make')
        local('docker build -t freekkalter/nginx .')
        local('docker push freekkalter/nginx')

    run('fleetctl destroy nginx.service')
    run('fleetctl start static/nginx.service')


def deploy_postgres():
    with lcd('postgres'):
        local('docker build -t freekkalter/postgres_km .')
        local('docker push freekkalter/postgres_km')

    # TODO: can not start 2 containers with same mounted data volume
    # postgres_port = get_running_port('postgres')
    # if postgres_port == '5432':
    #     new_port = '5433'
    # else:
    #     new_port = '5432'
    port = '5432'
    run('fleetctl stop postgres@{}.service'.format(port))
    run('fleetctl stop postgres-discovery@{}.service'.format(port))
    time.sleep(60)  # make sure postgres has shutdown gracefully

    run('fleetctl destroy postgres@{}.service'.format(port))
    run('fleetctl destroy postgres-discovery@{}.service'.format(port))

    run('fleetctl start templates/postgres@{}.service'.format(port))
    run('fleetctl start templates/postgres-discovery@{}.service'.format(port))


def get_running_port(service):
    output = run('fleetctl list-units')
    for line in output.split('\n'):
        match = re.match(''.join(['^', service, '@(\d{4})']), line)
        if match:
            return match.group(1)


def deploy_km():
    with lcd('km'):
        local('docker build -t freekkalter/km .')
        local('docker push freekkalter/km')
    run('docker pull freekkalter/km')
    port = get_running_port('km_webapp')
    if port == '4005':
        new_port = '4006'
    else:
        new_port = '4005'

    run('fleetctl start templates/km_webapp@{}.service'.format(new_port))
    run('fleetctl start templates/km_webapp-discovery@{}.service'.format(new_port))

    run('fleetctl destroy km_webapp@{}.service'.format(port))
    run('fleetctl destroy km_webapp-discovery@{}.service'.format(port))


# def backup_postgres():
def getSqlDump(directory):
    run('mkdir -p /home/core/backup')
    run('docker run -v /home/core/backup:/backup:rw --env POSTGRESADDRESS={}\
                    freekkalter/postgres_km /backup.sh'.format(get_postgres_address()[0]))
    get('/home/core/backup/backup.sql', directory)


def get_postgres_address():
    servers = run('etcdctl ls /services/postgres')
    serverlist = servers.split('\n')
    address = run('etcdctl get {}'.format(serverlist[0]))
    return (address.split(':'))


def backup():
    getSqlDump('.')
    # tar file and move to folder with format backup-{date}.sql
    local('mkdir -p ~/km-backup')
    local("tar -czf ~/km-backup/backup_`date +%d-%m-%Y.tar.gz` ./backup.sql")
    local("rm ./backup.sql")
