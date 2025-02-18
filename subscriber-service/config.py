import logging
import os
from typing import Dict


def read_algod_config_from_dir(directory: str) -> tuple[str, str]:
    """
    Reads the algod URL and token from the specified directory.
    """
    url_path = os.path.join(directory, "algod.net")
    try:
        with open(url_path, "r") as file:
            url = file.read().strip()
    except Exception as e:
        raise Exception(f"failed to read algod url: {e}")

    token_path = os.path.join(directory, "algod.token")
    try:
        with open(token_path, "r") as file:
            token = file.read().strip()
    except Exception as e:
        raise Exception(f"failed to read algod token: {e}")

    return "http://" + url, token

def devnet_algod_config() -> tuple[str, str]:
    """
    Returns the default algokit localnet algod URL and token.
    """
    return "http://localhost:4001", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"


def load_env(filename: str) -> Dict[str, str]:
    """
    Reads a set of key-value pairs from a file and returns them as a dictionary.

    Each line in the file can be in one of the following formats:
      - key=value
      - # comment
      - // comment
      - empty line

    Malformed lines (i.e. lines without an '=' separator) are logged as warnings and skipped.
    """
    env_map: Dict[str, str] = {}

    try:
        with open(filename, 'r') as file:
            for line in file:
                line = line.strip()

                # Skip empty lines and comments starting with '#' or '//'
                if not line or line.startswith('#') or line.startswith('//'):
                    continue

                parts = line.split('=', 1)
                if len(parts) != 2:
                    logging.warning(f"Malformed line in env file: {line}")
                    continue  # Skip malformed lines

                key = parts[0].strip()
                value = parts[1].strip()

                # Remove surrounding quotes if any
                value = value.strip('\'"')

                env_map[key] = value

    except Exception as e:
        raise Exception(f"Error reading file {filename}: {e}")

    return env_map