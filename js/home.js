function getCookie(name) {
   if (!document.cookie) {
      return null;
   }

   const xsrfCookies = document.cookie.split(';')
      .map(c => c.trim())
      .filter(c => c.startsWith(name + '='));

   if (xsrfCookies.length === 0) {
      return null;
   }

   return decodeURIComponent(xsrfCookies[0].split('=')[1]);
}

function contactRequest(ev) {
   ev && ev.preventDefault && ev.preventDefault()
   ev && ev.stopImmediatePropagation && ev.stopImmediatePropagation()

   var name = document.querySelector('input[name=name]').value
   var email = document.querySelector('input[name=email]').value
   var success = document.querySelector('div.alert-success')
   var failure = document.querySelector('div.alert-error')

   // Reset response holders
   success.innerHTML = ''
   failure.innerHTML = ''

   fetch("contact/request", {
      body: JSON.stringify({
         name: name,
         email: email
      }),
      headers: new Headers({
         'Content-Type': 'application/json',
         'X-XSRF-TOKEN': getCookie('_csrf')
      }),
      method: 'POST'
   })
   .then(function(resp) {

      if(resp.ok) {
         success.innerHTML = 'Request Sent'
         document.querySelector('form').reset()
         return true
      }
      return Promise.resolve(resp.json()).then(function(o) {
         failure.innerHTML = o.message
      })
   })
   .catch( function(err) {
      failure.innerHTML = 'network error'
   })
}

document.addEventListener("DOMContentLoaded", function() {
   document.querySelector('form').addEventListener('submit', contactRequest)
})
