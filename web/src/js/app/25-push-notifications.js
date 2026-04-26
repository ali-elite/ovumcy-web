  function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
      .replace(/\-/g, '+')
      .replace(/_/g, '/');

    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
      outputArray[i] = rawData.charCodeAt(i);
    }
    return outputArray;
  }

  function initPushSubscription() {
    var pushToggle = document.getElementById('push-notifications-toggle');
    if (!pushToggle) {
      return;
    }

    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
      pushToggle.parentElement.style.display = 'none';
      return;
    }

    var vapidPublicKey = document.head.querySelector('meta[name="vapid-public-key"]')?.content;
    if (!vapidPublicKey) {
      pushToggle.parentElement.style.display = 'none';
      return;
    }

    // Check current state
    navigator.serviceWorker.register('/sw.js', { scope: '/' }).then(function(reg) {
      return reg.pushManager.getSubscription();
    }).then(function(subscription) {
      if (subscription) {
        pushToggle.checked = true;
      }
    });

    pushToggle.addEventListener('change', function(e) {
      if (e.target.checked) {
        subscribeUser(vapidPublicKey, e.target);
      } else {
        unsubscribeUser(e.target);
      }
    });
  }

  function subscribeUser(vapidPublicKey, toggle) {
    navigator.serviceWorker.ready.then(function(reg) {
      return reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
      });
    }).then(function(subscription) {
      // Send to server
      return fetch('/api/settings/push/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': document.querySelector('input[name="csrf_token"]')?.value
        },
        body: JSON.stringify(subscription)
      });
    }).then(function(response) {
      if (!response.ok) {
        throw new Error('Failed to save subscription');
      }
    }).catch(function(err) {
      console.error('Failed to subscribe to push notifications', err);
      toggle.checked = false; // Revert
    });
  }

  function unsubscribeUser(toggle) {
    navigator.serviceWorker.ready.then(function(reg) {
      return reg.pushManager.getSubscription();
    }).then(function(subscription) {
      if (subscription) {
        return subscription.unsubscribe();
      }
    }).catch(function(err) {
      console.error('Error unsubscribing', err);
    });
  }
