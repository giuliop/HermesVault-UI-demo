"""
This script runs a service that asks the Algorand node for new transactions to the APP
and saves them in the database.

It saves eacn new deposit or change note to the txn_data.db sqlite database so that:
1) we can update the main frontend database with notes coming from other frontends (which we
   need to create merkle proofs for withdrawals)
2) we can clean up unconfirmed_notes in the main frontend database in (the rare) case the
   frontend crashes before the note is confirmed
"""
import json
import logging
import os

import algokit_subscriber as sub
import algosdk
import parse
from algokit_subscriber import AlgorandSubscriber, NamedTransactionFilter
from algokit_subscriber.types.subscription import TransactionFilter

import config
import db
from models import Deposit, Note, Withdrawal

# read configuration
env = config.load_env("../config/.env")
ALGOD_DIR = env["AlgodPath"]
APP_FILE = os.path.join(env["AppSetupDirPath"], "App.json")
DB_FILE = env["TxnsDbPath"]

# TODO: review logging
# logging.basicConfig(level=logging.INFO)
logging.getLogger().setLevel(logging.WARNING)
logger = logging.getLogger(__name__)

APP_ID = None
watermark = None

def get_watermark() -> int:
    return watermark

def set_watermark(new_watermark: int) -> None:
    global watermark
    watermark = new_watermark
    db.set_watermark(new_watermark)

def init():
    global APP_ID
    global watermark

    with open(APP_FILE) as f:
        app = json.load(f)
        APP_ID = app["id"]
        APP_CREATION_BLOCK = app["creationBlock"]

    db.initialize_db(DB_FILE)

    watermark = db.get_watermark()
    if watermark < APP_CREATION_BLOCK:
        set_watermark(APP_CREATION_BLOCK)

def handle_transaction(txn: sub.SubscribedTransaction, filter_name: str) -> None:
    """
    Process a deposit or withdrawal received from the subscriber
    """
    txn_id = txn.get("id")
    args = txn.get("application-transaction")["application-args"]
    result = txn.get("logs")[-1]
    leaf_index, tree_root = parse.txn_result(result)
    confirmed_block = txn.get("confirmed-round")

    if filter_name == parse.depositFilterName:
        commitment, address, amount = parse.deposit_args(args)
        note = Note(
            leaf_index=leaf_index,
            commitment=commitment,
            txn_id=txn_id,
        )
        deposit = Deposit(
            leaf_index=leaf_index,
            address=address,
            amount=amount,
        )
        db.retry(lambda: db.save_deposit(note, deposit, tree_root, confirmed_block))

    if filter_name == parse.withdrawFilterName:
        accounts = txn.get("application-transaction")["accounts"]
        commitment, address, nullifier, amount, fee = parse.withdraw_args(args, accounts)
        note = Note(
            leaf_index=leaf_index,
            commitment=commitment,
            txn_id=txn_id,
        )
        withdrawal = Withdrawal(
            leaf_index=leaf_index,
            address=accounts[1],
            nullifier=nullifier,
            amount=amount,
            fee=fee,
        )
        db.retry(lambda: db.save_withdrawal(note, withdrawal, tree_root, confirmed_block))


def main():
    init()

    if ALGOD_DIR == "":
        algod_address, algod_token = config.devnet_algod_config()
    else:
        algod_address, algod_token = config.read_algod_config_from_dir(ALGOD_DIR)

    algod = algosdk.v2client.algod.AlgodClient(algod_token, algod_address)

    depositFilter = NamedTransactionFilter(
        name=parse.depositFilterName,
        filter=TransactionFilter(app_id=APP_ID, method_signature=parse.DEPOSIT_SIGNATURE),
    )
    withdrawFilter = NamedTransactionFilter(
        name=parse.withdrawFilterName,
        filter=TransactionFilter(app_id=APP_ID, method_signature=parse.WITHDRAW_SIGNATURE),
    )
    subscriber = AlgorandSubscriber(
        config={
            "filters": [depositFilter, withdrawFilter],
            "sync_behaviour": "sync-oldest",
            "watermark_persistence": {"get": get_watermark, "set": set_watermark},
            "frequency_in_seconds": 5,
        },
        algod_client=algod,
    )

    logger.info("Starting subscriber for app_id %d...", APP_ID)

    subscriber.on(depositFilter['name'], handle_transaction)
    subscriber.on(withdrawFilter['name'], handle_transaction)

    def handle_error(error: Exception, _) -> None:
        logger.error("Subscriber error: %s", error)
    subscriber.on_error(handle_error)

    subscriber.start()


if __name__ == "__main__":
    main()
