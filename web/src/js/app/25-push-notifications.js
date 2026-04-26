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
    var status = document.querySelector('[data-push-status]');

    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
      pushToggle.parentElement.style.display = 'none';
      setPushStatus(status, 'error');
      return;
    }

    var vapidPublicKey = document.head.querySelector('meta[name="vapid-public-key"]')?.content;
    if (!vapidPublicKey) {
      pushToggle.parentElement.style.display = 'none';
      setPushStatus(status, 'error');
      return;
    }

    // Check current state
    navigator.serviceWorker.register('/sw.js', { scope: '/' }).then(function(reg) {
      return reg.pushManager.getSubscription();
    }).then(function(subscription) {
      if (subscription) {
        pushToggle.checked = true;
        return savePushSubscription(subscription).then(function() {
          setPushStatus(status, 'enabled');
        });
      }
      setPushStatus(status, 'ready');
    }).catch(function(err) {
      console.error('Failed to initialize push notifications', err);
      setPushStatus(status, 'error');
    });

    pushToggle.addEventListener('change', function(e) {
      if (e.target.checked) {
        subscribeUser(vapidPublicKey, e.target, status);
      } else {
        unsubscribeUser(e.target, status);
      }
    });
  }

  function setPushStatus(status, key) {
    if (!status || !status.dataset) {
      return;
    }
    var attr = 'message' + String(key || '').replace(/(^|_)([a-z])/g, function(_, __, char) {
      return char.toUpperCase();
    });
    status.textContent = status.dataset[attr] || '';
    status.classList.toggle('status-error', key === 'error' || key === 'denied');
    status.classList.toggle('journal-muted', key !== 'error' && key !== 'denied');
  }

  function csrfTokenValue() {
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (meta && meta.content) {
      return meta.content;
    }
    var input = document.querySelector('input[name="csrf_token"]');
    return input ? input.value : '';
  }

  function pushSubscriptionPayload(subscription) {
    var serialized = typeof subscription.toJSON === 'function' ? subscription.toJSON() : subscription;
    var keys = serialized && serialized.keys ? serialized.keys : {};
    var body = new URLSearchParams();
    body.set('csrf_token', csrfTokenValue());
    body.set('endpoint', serialized && serialized.endpoint ? serialized.endpoint : '');
    body.set('p256dh', keys.p256dh || '');
    body.set('auth', keys.auth || '');
    return body;
  }

  function savePushSubscription(subscription) {
    return fetch('/api/settings/push/subscribe', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Accept': 'application/json'
      },
      body: pushSubscriptionPayload(subscription)
    }).then(function(response) {
      if (!response.ok) {
        throw new Error('Failed to save subscription');
      }
      return response;
    });
  }

  function removePushSubscription(subscription) {
    var body = new URLSearchParams();
    body.set('csrf_token', csrfTokenValue());
    body.set('endpoint', subscription && subscription.endpoint ? subscription.endpoint : '');
    return fetch('/api/settings/push/unsubscribe', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Accept': 'application/json'
      },
      body: body
    }).then(function(response) {
      if (!response.ok) {
        throw new Error('Failed to remove subscription');
      }
      return response;
    });
  }

  function subscribeUser(vapidPublicKey, toggle, status) {
    toggle.disabled = true;
    setPushStatus(status, 'enabling');
    navigator.serviceWorker.ready.then(function(reg) {
      return reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
      });
    }).then(function(subscription) {
      return savePushSubscription(subscription);
    }).then(function() {
      toggle.checked = true;
      setPushStatus(status, 'enabled');
    }).catch(function(err) {
      console.error('Failed to subscribe to push notifications', err);
      toggle.checked = false;
      setPushStatus(status, typeof Notification !== 'undefined' && Notification.permission === 'denied' ? 'denied' : 'error');
    }).finally(function() {
      toggle.disabled = false;
    });
  }

  function unsubscribeUser(toggle, status) {
    toggle.disabled = true;
    navigator.serviceWorker.ready.then(function(reg) {
      return reg.pushManager.getSubscription();
    }).then(function(subscription) {
      if (subscription) {
        return removePushSubscription(subscription).then(function() {
          return subscription.unsubscribe();
        });
      }
    }).then(function() {
      toggle.checked = false;
      setPushStatus(status, 'disabled');
    }).catch(function(err) {
      console.error('Error unsubscribing', err);
      toggle.checked = true;
      setPushStatus(status, 'error');
    }).finally(function() {
      toggle.disabled = false;
    });
  }
