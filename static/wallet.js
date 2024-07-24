import { PeraWalletConnect } from './pera-connect.bundle.js';

const peraWallet = new PeraWalletConnect();
let accountAddress = "";

function updateUI(accounts) {
    const addressInput = document.querySelector('[data-wallet-address]');
    const depositButton = document.querySelector('[data-wallet-deposit-button]');
    const walletButton = document.querySelector('[data-wallet-connect-button]');
    if (accounts.length) {
        accountAddress = accounts[0];
        addressInput.value = accountAddress;
        // Trigger the blur event to trim the address in the UI
        addressInput.dispatchEvent(new Event('blur'));
        depositButton.classList.remove('hidden');
        walletButton.textContent = "Disconnect Wallet";
    } else {
        accountAddress = "";
        addressInput.value = "";
        depositButton.classList.add('hidden');
        walletButton.textContent = "Connect Wallet";
    }
}

function reconnectSession() {
    // Reconnect to the session when the component is mounted
    peraWallet
        .reconnectSession()
        .then((accounts) => {
            peraWallet.connector?.on("disconnect", handleDisconnectWalletClick);
            updateUI(accounts);
        })
        .catch((e) => console.log(e));
}

function handleConnectWalletClick(event) {
    event.preventDefault();

    peraWallet
        .connect()
        .then((accounts) => {
            peraWallet.connector?.on("disconnect", handleDisconnectWalletClick);
            updateUI(accounts);
        })
        .catch((error) => {
            if (error?.data?.type !== "CONNECT_MODAL_CLOSED") {
                console.log(error);
            }
        });
}

function handleDisconnectWalletClick(event) {
    event.preventDefault();

    peraWallet.disconnect().catch((error) => {
        console.log(error);
    });

    updateUI([]);
}

// trigger on connet wallet button and confirm deposit button
document.addEventListener('click', async (event) => {
    if (event.target.matches('[data-wallet-connect-button]')) {
        event.preventDefault();
        if (accountAddress) {
            handleDisconnectWalletClick(event);
        } else {
            handleConnectWalletClick(event);
        }
    }
    if (event.target.matches('[data-wallet-confirm-deposit-button]')) {
        event.preventDefault();
        let data = [
            { data: "Deposit transaction",
              message: "This is just a demo so we sign dummy data, in the real application we'll sign the deposit transaction"
            }
        ]
        try {
            const signed = await peraWallet.signData(data, accountAddress)
            console.log(signed);
            document.querySelector('[data-wallet-signedTxn-input]').value = (
                 signed);
            const form = event.target.closest('form');
            htmx.trigger(form, 'submit');
        } catch (error) {
            console.log(error);
            let errorBox = document.querySelector('[data-wallet-errorBox]');
            errorBox.innerHTML = (
                "Error signing the transaction, please try again");
            htmx.trigger(errorBox, 'htmx:after-swap');
        }
    };
});

// trigger on wallet form load
document.addEventListener('htmx:load', (event) => {
    if (event.detail.elt.matches('[data-wallet]')) {
        if (!accountAddress) {
            reconnectSession();
        } else {
            updateUI([accountAddress]);
        }
    }
});

// Make functions and variables accessible from the console for debugging
// window.peraWallet = peraWallet;
// window.accountAddress = accountAddress;
// window.updateUIAfterConnect = updateUI;
// window.reconnectSession = reconnectSession;
// window.handleConnectWalletClick = handleConnectWalletClick;
// window.handleDisconnectWalletClick = handleDisconnectWalletClick;