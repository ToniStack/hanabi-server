/*
 *  Javascript for the Hanabi client
 */

/*
 *  Auth0 initialization
 */

var lock = new Auth0Lock(
    'XrDVzvVvlWonCa9U9RPjW5182r50Ff2E',
    'isaacserver.auth0.com'
);
var userProfile;
var userToken = localStorage.getItem('userToken');

lock.on("authenticated", function(authResult) {
    // Use the token in authResult to getProfile() and save it to localStorage
    lock.getProfile(authResult.idToken, function(error, profile) {
        if (error) {
            showError('Failed to get the user\'s profile: ' + error.message);
        }

        localStorage.setItem('userToken', authResult.idToken);
        userProfile = profile;
    });
});

// They are logged in
if (userToken) {
    // Hide the login
    $('#banner').hide();
    $('#login').hide();

    // Get the profile
    lock.getProfile(userToken, function (error, profile) {
        if (error) {
            showError('Failed to get the user\'s profile: ' + error.message);
        }        
        userProfile = profile;
    });

    var socket = new WebSocket('ws://' + window.location.host + '/ws')
    socket.onmessage = function(event) {
        display(JSON.parse(event.data))
    }
    console.log('logged in');

    

// They are not logged in
} else {
    
}

/*
 *  Login stuff
 */

$('#login').click(function(e) {
    e.preventDefault();
    lock.show();
});

/*
 *   Miscellaneous functions
 */

function showError(message) {
    $('#login').hide();
    $('#lobby').hide();
    $('#hanabi').hide();
    $('#banner').show();
    $('#error').show();
    $('#error-message').html(message);
}
