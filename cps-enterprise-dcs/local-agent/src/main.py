"""
Local Agent Entry Point
=======================
Main entry point for the CP'S Dynamics Commerce System AgenticOS Local Agent.

Usage:
    python -m local_agent.main --branch-id BR001 --region R001

Environment Variables:
    DCS_AGENT_ID: Unique agent identifier
    DCS_BRANCH_ID: Branch identifier
    DCS_REGION_ID: Region identifier
    DCS_DB_PATH: Path to SQLite database
    DCS_MASTER_KEY: Encryption master key (base64 encoded)
    DCS_SYNC_INTERVAL: Sync interval in seconds
    DCS_GRPC_PORT: gRPC server port
"""

import asyncio
import os
import sys
import base64
import signal
from pathlib import Path

import click
from dotenv import load_dotenv

from .agent import LocalAgent, AgentConfig, AgentState
from .grpc_server import LocalAgentGRPCServer


# Load environment variables
load_dotenv()


async def run_agent(config: AgentConfig):
    """Run the local agent."""
    # Create agent
    agent = LocalAgent(config)
    
    # Create gRPC server
    grpc_server = LocalAgentGRPCServer(agent, port=config.pos_interface_port)
    
    # Setup signal handlers
    shutdown_event = asyncio.Event()
    
    def signal_handler(sig, frame):
        print("\nShutdown signal received...")
        shutdown_event.set()
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        # Initialize agent
        await agent.initialize()
        
        # Start gRPC server
        await grpc_server.start()
        
        print(f"""
╔══════════════════════════════════════════════════════════════╗
║     CP'S Dynamics Commerce System AgenticOS - Local Agent    ║
║                                                              ║
║  Agent ID:  {config.agent_id:<45} ║
║  Branch:    {config.branch_id:<45} ║
║  Region:    {config.region_id:<45} ║
║  State:     {agent.state.value:<45} ║
║  gRPC Port: {config.pos_interface_port:<45} ║
║                                                              ║
║  Press Ctrl+C to shutdown                                    ║
╚══════════════════════════════════════════════════════════════╝
        """)
        
        # Wait for shutdown signal
        await shutdown_event.wait()
        
    finally:
        # Cleanup
        await grpc_server.stop()
        await agent.shutdown()
        print("Agent shutdown complete")


@click.command()
@click.option('--agent-id', default=lambda: os.getenv('DCS_AGENT_ID', None),
              help='Unique agent identifier')
@click.option('--branch-id', default=lambda: os.getenv('DCS_BRANCH_ID', None),
              help='Branch identifier')
@click.option('--region-id', default=lambda: os.getenv('DCS_REGION_ID', None),
              help='Region identifier')
@click.option('--db-path', default=lambda: os.getenv('DCS_DB_PATH', 'events.db'),
              help='Path to SQLite database')
@click.option('--sync-interval', default=lambda: int(os.getenv('DCS_SYNC_INTERVAL', '30')),
              type=int, help='Sync interval in seconds')
@click.option('--grpc-port', default=lambda: int(os.getenv('DCS_GRPC_PORT', '50051')),
              type=int, help='gRPC server port')
@click.option('--regional-endpoint', default=lambda: os.getenv('DCS_REGIONAL_ENDPOINT', None),
              help='Regional agent endpoint')
@click.option('--master-key', default=lambda: os.getenv('DCS_MASTER_KEY', None),
              help='Encryption master key (base64 encoded)')
def main(
    agent_id: str,
    branch_id: str,
    region_id: str,
    db_path: str,
    sync_interval: int,
    grpc_port: int,
    regional_endpoint: str,
    master_key: str
):
    """Run the CP'S Enterprise DCS Local Agent."""
    
    # Validate required parameters
    if not agent_id:
        agent_id = f"agent-{base64.urlsafe_b64encode(os.urandom(6)).decode()[:8]}"
        print(f"Generated agent ID: {agent_id}")
    
    if not branch_id:
        print("Error: --branch-id is required", file=sys.stderr)
        sys.exit(1)
    
    if not region_id:
        region_id = "default"
    
    # Decode master key if provided
    decoded_key = None
    if master_key:
        try:
            decoded_key = base64.b64decode(master_key)
        except Exception as e:
            print(f"Error decoding master key: {e}", file=sys.stderr)
            sys.exit(1)
    
    # Create configuration
    config = AgentConfig(
        agent_id=agent_id,
        branch_id=branch_id,
        region_id=region_id,
        db_path=db_path,
        sync_interval_seconds=sync_interval,
        batch_size=100,
        enable_encryption=True,
        master_key=decoded_key,
        regional_agent_endpoint=regional_endpoint,
        pos_interface_port=grpc_port
    )
    
    # Run agent
    try:
        asyncio.run(run_agent(config))
    except KeyboardInterrupt:
        print("\nInterrupted by user")
    except Exception as e:
        print(f"Fatal error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
