/*
 *  HANABI JAVASCRIPT - LOGIN SCREEN
 */

 /*
  *  Auth0 initialization
  */

 var lock = new Auth0Lock(
     'XrDVzvVvlWonCa9U9RPjW5182r50Ff2E',
     'isaacserver.auth0.com',
     {
         auth: {
             redirectUrl: window.location.href + 'login',
             responseType: 'code',
             params: {
                 scope: 'openid email'
             }
         }
     }
 );

 /*
  *  Login stuff
  */

 $('#login').click(function(e) {
     lock.show();
 });
