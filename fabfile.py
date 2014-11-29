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
    postgres_port = get_running_postgres_port()

    if postgres_port == '5432':
        new_port = '5433'
    else:
        new_port = '5432'
    run('fleetctl start templates/postgres@{}.service'.format(new_port))
    run(
        'fleetctl start templates/postgres-discovery@{}.service'.format(new_port))

    run('fleetctl destroy postgres@{}.service'.format(postgres_port))
    run('fleetctl destroy postgres-discovery@{}.service'.format(postgres_port))


def get_running_postgres_port():
    output = run('fleetctl list-units')
    for line in output.split('\n'):
        match = re.match('^postgres@(\d{4})', line)
        if match:
            return match.group(1)
