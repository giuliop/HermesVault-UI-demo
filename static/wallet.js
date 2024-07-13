import { PeraWalletConnect } from './pera-connect.bundle.js';

const peraWallet = new PeraWalletConnect();
let accountAddress = "";

function updateUI(accounts) {
    if (accounts.length) {
        accountAddress = accounts[0];
        document.querySelector('[data-wallet-address] input').value = accountAddress;
        document.querySelector('form').hidden = false;
        document.querySelector('[data-wallet-button]').textContent = "Disconnect Wallet";
    } else {
        accountAddress = "";
        document.querySelector('[data-wallet-address] input').value = "";
        document.querySelector('form').hidden = true;
        document.querySelector('[data-wallet-button]').textContent = "Connect Pera Wallet";
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

// Use event delegation for the wallet button
document.addEventListener('click', (event) => {
    if (event.target.matches('[data-wallet-button]')) {
        event.preventDefault();
        if (accountAddress) {
            handleDisconnectWalletClick(event);
        } else {
            handleConnectWalletClick(event);
        }
    }
});

// Listen for htmx:afterSwap
document.getElementById('tabs').addEventListener('htmx:afterSwap', (event) => {
    const depositTab = event.detail.target.querySelector(
        '#tab-content[data-wallet]');
    if (depositTab) {
        if (!accountAddress) {
            reconnectSession();
        } else {
            updateUI([accountAddress]);
        }
    }
});


// Make functions and variables accessible from the console
window.peraWallet = peraWallet;
window.accountAddress = accountAddress;
window.updateUIAfterConnect = updateUI;
window.reconnectSession = reconnectSession;
window.handleConnectWalletClick = handleConnectWalletClick;
window.handleDisconnectWalletClick = handleDisconnectWalletClick;