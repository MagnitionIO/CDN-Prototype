import yaml
import os
import subprocess
from jinja2 import Template

def load_yaml_config(file_path):
    with open(file_path, 'r') as stream:
        return yaml.safe_load(stream)

def generate_vcl_file(vcl_template, output_path):
    with open(output_path, 'w') as vcl_file:
        vcl_file.write(vcl_template)

def generate_docker_compose(yaml_config, output_path):
    template = Template('''
version: '3.7'
services:
  {% for i in range(1, varnish['l1']['count'] + 1) %}
  varnish_l1_{{ i }}:
    image: varnish:latest
    volumes:
      - ./{{ varnish['l1']['vcl_file'] }}:/etc/varnish/default.vcl
    environment:
      - VARNISH_STORAGE={{ varnish['l1']['storage'] }},{{ varnish['l1']['storage_size'] }}
      - VARNISH_SIZE={{ varnish['l1']['max_object_size'] }}
    ports:
      - "{{ 8080 + i }}:80"
  {% endfor %}
  {% for i in range(1, varnish['l2']['count'] + 1) %}
  varnish_l2_{{ i }}:
    image: varnish:latest
    volumes:
      - ./{{ varnish['l2']['vcl_file'] }}:/etc/varnish/default.vcl
    environment:
      - VARNISH_STORAGE={{ varnish['l2']['storage'] }},{{ varnish['l2']['storage_size'] }}
      - VARNISH_SIZE={{ varnish['l2']['max_object_size'] }}
    ports:
      - "{{ 8080 + varnish['l1']['count'] + i }}:80"
  {% endfor %}
    ''')

    with open(output_path, 'w') as compose_file:
        compose_file.write(template.render(varnish=yaml_config['varnish']))

def run_automation(yaml_file_path):
    config = load_yaml_config(yaml_file_path)
    
    l1_vcl_template = "vcl 4.0; /* L1 VCL Configurations */"
    l2_vcl_template = "vcl 4.0; /* L2 VCL Configurations */"
    
    generate_vcl_file(l1_vcl_template, config['varnish']['l1']['vcl_file'])
    generate_vcl_file(l2_vcl_template, config['varnish']['l2']['vcl_file'])
    
    generate_docker_compose(config, 'docker-compose.yml')
    
    subprocess.run(['docker-compose', 'up', '-d'])
    
    done_file_path = '/done'
    while not os.path.exists(done_file_path):
        pass
    
    subprocess.run(['docker-compose', 'down'])

if __name__ == "__main__":
    yaml_file_path = 'config.yaml'
    run_automation(yaml_file_path)
