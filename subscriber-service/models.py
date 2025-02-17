from dataclasses import dataclass


def format_bytes(b: bytes) -> str:
    """Format a bytes object as hex with a space between each byte."""
    return " ".join(f"{byte:02x}" for byte in b)

@dataclass(frozen=True)
class Note:
    leaf_index: int
    commitment: bytes
    txn_id: str

    def __repr__(self):
        return (
            f"Note(leaf_index={self.leaf_index!r}, "
            f"commitment={format_bytes(self.commitment)}, "
            f"txn_id={self.txn_id!r})"
        )

@dataclass(frozen=True)
class Deposit:
    leaf_index: int
    address: str
    amount: int

    def __repr__(self):
        return (
            f"Deposit(leaf_index={self.leaf_index!r}, "
            f"address={self.address!r}, "
            f"amount={self.amount!r})"
        )

@dataclass(frozen=True)
class Withdrawal:
    leaf_index: int
    address: str
    nullifier: bytes
    amount: int
    fee: int

    def __repr__(self):
        return (
            f"Withdrawal(leaf_index={self.leaf_index!r}, "
            f"address={self.address!r}, "
            f"nullifier={format_bytes(self.nullifier)}, "
            f"amount={self.amount!r}, "
            f"fee={self.fee!r})"
        )