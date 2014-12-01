from fabric.api import *
import re

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
