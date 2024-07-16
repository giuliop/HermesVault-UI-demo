import { PeraWalletConnect } from './pera-connect.bundle.js';

const peraWallet = new PeraWalletConnect();
let accountAddress = "";

function updateUI(accounts) {
    const addressInput = document.querySelector('[data-wallet-address]');
    const submitButton = document.querySelector('[data-wallet-submit-button]');
    const walletButton = document.querySelector('[data-wallet-connect-button]');
    if (accounts.length) {
        accountAddress = accounts[0];
        addressInput.value = accountAddress;
        // Trigger the blur event to trim the address in the UI
        addressInput.dispatchEvent(new Event('blur'));
        submitButton.classList.remove('hidden');
        walletButton.textContent = "Disconnect Wallet";
    } else {
        accountAddress = "";
        addressInput.value = "";
        submitButton.classList.add('hidden');
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

// trigger on wallet button click
document.addEventListener('click', (event) => {
    if (event.target.matches('[data-wallet-connect-button]')) {
        event.preventDefault();
        if (accountAddress) {
            handleDisconnectWalletClick(event);
        } else {
            handleConnectWalletClick(event);
        }
    }
});

// trigger on wallet form load
document.body.addEventListener('htmx:load', (event) => {
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