#!/usr/bin/env python3
"""
Mock SLURM REST API Server for Integration Testing
Simulates rocky.ar.jontk.com SLURM cluster responses for testing when the real cluster is unavailable
"""

import json
import time
import random
from datetime import datetime, timedelta
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs
import ssl


class MockSLURMHandler(BaseHTTPRequestHandler):
    """Handler for mock SLURM REST API endpoints"""
    
    def __init__(self, *args, **kwargs):
        self.jwt_secret = "test-secret"
        self.valid_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI2NTQzMDU1NzksImlhdCI6MTc1NDMwNTU3OSwic3VuIjoicm9vdCJ9.ECpVcCqD1eXUk5-3LZLCDVDolvCJkhGNj24IvqwCkPs"
        super().__init__(*args, **kwargs)
    
    def log_message(self, format, *args):
        """Override to reduce noise in tests"""
        return
    
    def do_GET(self):
        """Handle GET requests"""
        parsed_url = urlparse(self.path)
        path = parsed_url.path
        query = parse_qs(parsed_url.query)
        
        # Check authentication
        if not self._check_auth():
            return
        
        # Route requests
        if path == "/slurm/v0.0.43/ping":
            self._handle_ping()
        elif path == "/slurm/v0.0.43/diag":
            self._handle_diag()
        elif path == "/slurm/v0.0.43/jobs":
            self._handle_jobs(query)
        elif path == "/slurm/v0.0.43/nodes":
            self._handle_nodes(query)
        elif path == "/slurm/v0.0.43/partitions":
            self._handle_partitions(query)
        elif path == "/slurm/v0.0.43/licenses":
            self._handle_licenses()
        elif path == "/slurm/v0.0.43/reservations":
            self._handle_reservations()
        else:
            self._send_error(404, "Not Found")
    
    def _check_auth(self):
        """Check JWT authentication"""
        auth_header = self.headers.get('X-SLURM-USER-TOKEN')
        if not auth_header:
            self._send_error(401, "No authentication token provided")
            return False
        
        if auth_header != self.valid_token:
            self._send_error(401, "Invalid authentication token")
            return False
        
        return True
    
    def _send_json_response(self, data, status=200):
        """Send JSON response"""
        self.send_response(status)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(data, indent=2).encode())
    
    def _send_error(self, status, message):
        """Send error response"""
        self.send_response(status)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        error_response = {
            "errors": [{"error": message}],
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            }
        }
        self.wfile.write(json.dumps(error_response).encode())
    
    def _handle_ping(self):
        """Handle ping endpoint"""
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "pings": [
                {
                    "hostname": "rocky.ar.jontk.com",
                    "ping": "UP",
                    "status": 0,
                    "mode": "primary"
                }
            ]
        }
        self._send_json_response(response)
    
    def _handle_diag(self):
        """Handle diagnostics endpoint"""
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "statistics": {
                "cluster": "rocky-cluster",
                "controller": {
                    "host": "rocky.ar.jontk.com",
                    "version": "23.11.0",
                    "started": int(time.time() - 86400),  # Started 1 day ago
                    "uptime": 86400
                },
                "nodes": {
                    "total": 8,
                    "idle": 2,
                    "allocated": 4,
                    "down": 1,
                    "mixed": 1
                },
                "jobs": {
                    "total": 15,
                    "running": 8,
                    "pending": 5,
                    "completed": 2
                },
                "users": {
                    "active": 6,
                    "total": 12
                }
            }
        }
        self._send_json_response(response)
    
    def _handle_jobs(self, query):
        """Handle jobs endpoint"""
        limit = int(query.get('limit', [100])[0])
        
        # Generate mock job data
        jobs = []
        job_states = ['RUNNING', 'PENDING', 'COMPLETED', 'CANCELLED', 'FAILED', 'TIMEOUT']
        partitions = ['gpu', 'cpu', 'bigmem', 'debug']
        users = ['user1', 'user2', 'user3', 'admin', 'test_user']
        
        for i in range(min(limit, 15)):
            job_id = 1000 + i
            state = random.choice(job_states)
            start_time = int(time.time()) - random.randint(0, 86400)
            
            job = {
                "job_id": job_id,
                "array_job_id": 0,
                "array_task_id": None,
                "name": f"test_job_{i}",
                "user_id": random.randint(1000, 1010),
                "user_name": random.choice(users),
                "group_id": random.randint(100, 110),
                "state": state,
                "exit_code": 0 if state == 'COMPLETED' else None,
                "submit_time": start_time - random.randint(0, 3600),
                "start_time": start_time if state in ['RUNNING', 'COMPLETED'] else 0,
                "end_time": start_time + random.randint(1800, 7200) if state == 'COMPLETED' else 0,
                "partition": random.choice(partitions),
                "nodes": f"node{random.randint(1, 8):02d}" if state == 'RUNNING' else "",
                "node_count": random.randint(1, 4) if state in ['RUNNING', 'COMPLETED'] else 0,
                "cpu_count": random.randint(1, 32),
                "tasks": random.randint(1, 8),
                "time_limit": random.randint(60, 1440),  # 1 hour to 24 hours
                "working_directory": f"/home/{random.choice(users)}/jobs",
                "standard_output": f"/home/{random.choice(users)}/logs/job_{job_id}.out",
                "standard_error": f"/home/{random.choice(users)}/logs/job_{job_id}.err",
                "priority": random.randint(1, 100),
                "qos": "normal",
                "account": "default",
                "command": f"./test_script_{i}.sh",
                "tres_req": {
                    "cpu": random.randint(1, 32),
                    "mem": f"{random.randint(1, 64)}G",
                    "node": random.randint(1, 4)
                }
            }
            jobs.append(job)
        
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "jobs": jobs
        }
        self._send_json_response(response)
    
    def _handle_nodes(self, query):
        """Handle nodes endpoint"""
        nodes = []
        node_states = ['IDLE', 'ALLOCATED', 'MIXED', 'DOWN', 'DRAIN']
        
        for i in range(1, 9):  # 8 nodes
            node_name = f"node{i:02d}"
            state = random.choice(node_states)
            
            node = {
                "name": node_name,
                "hostname": f"{node_name}.rocky.ar.jontk.com",
                "state": state,
                "cpus": random.choice([16, 32, 64]),
                "cpus_allocated": random.randint(0, 32) if state in ['ALLOCATED', 'MIXED'] else 0,
                "real_memory": random.choice([64000, 128000, 256000]),  # MB
                "alloc_memory": random.randint(0, 128000) if state in ['ALLOCATED', 'MIXED'] else 0,
                "free_memory": random.randint(32000, 256000),
                "tmp_disk": 1000000,  # MB
                "weight": 1,
                "features": ["gpu", "infiniband"] if i <= 4 else ["cpu"],
                "architecture": "x86_64",
                "operating_system": "Linux 5.15.0",
                "boot_time": int(time.time() - random.randint(86400, 604800)),
                "slurmd_start_time": int(time.time() - random.randint(3600, 86400)),
                "last_busy": int(time.time() - random.randint(0, 3600)),
                "partitions": ["gpu", "cpu"] if i <= 4 else ["cpu", "bigmem"],
                "alloc_cpus": random.randint(0, 32) if state in ['ALLOCATED', 'MIXED'] else 0,
                "idle_cpus": random.randint(0, 32),
                "tres": f"cpu={random.choice([16, 32, 64])},mem={random.choice([64, 128, 256])}G,gres/gpu={random.randint(0, 4) if i <= 4 else 0}",
                "reason": "none" if state != 'DOWN' else "Node not responding",
                "reason_time": int(time.time()) if state == 'DOWN' else None,
                "reason_user": "slurm" if state == 'DOWN' else None
            }
            nodes.append(node)
        
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "nodes": nodes
        }
        self._send_json_response(response)
    
    def _handle_partitions(self, query):
        """Handle partitions endpoint"""
        partitions = [
            {
                "name": "gpu",
                "nodes": ["node01", "node02", "node03", "node04"],
                "total_nodes": 4,
                "total_cpus": 128,
                "allocated_nodes": 2,
                "idle_nodes": 1,
                "down_nodes": 1,
                "state": "UP",
                "max_time": 1440,  # 24 hours
                "default_time": 60,
                "max_nodes": 4,
                "max_cpus_per_node": 32,
                "default_memory_per_cpu": 4000,
                "max_memory_per_cpu": 8000,
                "priority_tier": 1,
                "allow_accounts": "all",
                "deny_accounts": "",
                "allow_groups": "all",
                "deny_groups": "",
                "allow_qos": "normal,high",
                "deny_qos": "",
                "tres": "cpu=128,mem=512G,gres/gpu=16"
            },
            {
                "name": "cpu",
                "nodes": ["node01", "node02", "node03", "node04", "node05", "node06", "node07", "node08"],
                "total_nodes": 8,
                "total_cpus": 256,
                "allocated_nodes": 4,
                "idle_nodes": 2,
                "down_nodes": 2,
                "state": "UP",
                "max_time": 2880,  # 48 hours
                "default_time": 120,
                "max_nodes": 8,
                "max_cpus_per_node": 64,
                "default_memory_per_cpu": 2000,
                "max_memory_per_cpu": 4000,
                "priority_tier": 2,
                "allow_accounts": "all",
                "deny_accounts": "",
                "allow_groups": "all",
                "deny_groups": "",
                "allow_qos": "normal",
                "deny_qos": "high",
                "tres": "cpu=256,mem=1024G"
            },
            {
                "name": "bigmem",
                "nodes": ["node07", "node08"],
                "total_nodes": 2,
                "total_cpus": 128,
                "allocated_nodes": 1,
                "idle_nodes": 1,
                "down_nodes": 0,
                "state": "UP",
                "max_time": 4320,  # 72 hours
                "default_time": 240,
                "max_nodes": 2,
                "max_cpus_per_node": 64,
                "default_memory_per_cpu": 8000,
                "max_memory_per_cpu": 16000,
                "priority_tier": 1,
                "allow_accounts": "privileged,admin",
                "deny_accounts": "",
                "allow_groups": "research",
                "deny_groups": "",
                "allow_qos": "normal,high",
                "deny_qos": "",
                "tres": "cpu=128,mem=2048G"
            },
            {
                "name": "debug",
                "nodes": ["node01", "node02"],
                "total_nodes": 2,
                "total_cpus": 64,
                "allocated_nodes": 0,
                "idle_nodes": 2,
                "down_nodes": 0,
                "state": "UP",
                "max_time": 30,  # 30 minutes
                "default_time": 10,
                "max_nodes": 1,
                "max_cpus_per_node": 8,
                "default_memory_per_cpu": 1000,
                "max_memory_per_cpu": 2000,
                "priority_tier": 1,
                "allow_accounts": "all",
                "deny_accounts": "",
                "allow_groups": "all",
                "deny_groups": "",
                "allow_qos": "normal",
                "deny_qos": "",
                "tres": "cpu=64,mem=128G"
            }
        ]
        
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "partitions": partitions
        }
        self._send_json_response(response)
    
    def _handle_licenses(self):
        """Handle licenses endpoint"""
        licenses = [
            {
                "name": "matlab",
                "total": 10,
                "used": 3,
                "available": 7,
                "remote": False
            },
            {
                "name": "ansys",
                "total": 5,
                "used": 2,
                "available": 3,
                "remote": False
            }
        ]
        
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "licenses": licenses
        }
        self._send_json_response(response)
    
    def _handle_reservations(self):
        """Handle reservations endpoint"""
        reservations = [
            {
                "name": "maintenance_window",
                "start_time": int(time.time() + 86400),  # Tomorrow
                "end_time": int(time.time() + 90000),    # Tomorrow + 1 hour
                "duration": 3600,
                "nodes": ["node07", "node08"],
                "node_count": 2,
                "users": ["admin"],
                "accounts": ["maintenance"],
                "state": "INACTIVE",
                "flags": ["MAINT"]
            }
        ]
        
        response = {
            "meta": {
                "plugin": {"type": "openapi/v0.0.43", "name": "REST v0.0.43"},
                "Slurm": {"version": {"major": 23, "micro": 0, "minor": 11}, "release": "23.11.0"}
            },
            "errors": [],
            "reservations": reservations
        }
        self._send_json_response(response)


def run_mock_server(port=6820, use_ssl=True):
    """Run the mock SLURM server"""
    server_address = ('', port)
    httpd = HTTPServer(server_address, MockSLURMHandler)
    
    if use_ssl:
        # Create a self-signed certificate for testing
        import ssl
        context = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
        context.load_cert_chain('test-cert.pem', 'test-key.pem')
        httpd.socket = context.wrap_socket(httpd.socket, server_side=True)
        protocol = "HTTPS"
    else:
        protocol = "HTTP"
    
    print(f"Mock SLURM server running on {protocol} port {port}")
    print(f"Simulating rocky.ar.jontk.com cluster")
    print(f"Valid JWT token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjI2NTQzMDU1NzksImlhdCI6MTc1NDMwNTU3OSwic3VuIjoicm9vdCJ9.ECpVcCqD1eXUk5-3LZLCDVDolvCJkhGNj24IvqwCkPs")
    print("Press Ctrl+C to stop")
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down mock server...")
        httpd.shutdown()


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Mock SLURM REST API Server")
    parser.add_argument("--port", type=int, default=6820, help="Port to run on (default: 6820)")
    parser.add_argument("--no-ssl", action="store_true", help="Disable SSL/TLS")
    
    args = parser.parse_args()
    
    run_mock_server(port=args.port, use_ssl=not args.no_ssl)