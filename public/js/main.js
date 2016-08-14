/*
 *  HANABI JAVASCRIPT - MAIN
 */

/*
TODO:
- create game ruleset should remember from last time, ask from server
*/

/*
 *  Global variables
 */

var currentGames = [];
var roomList = {};
var createGameOptions = {
    'name':    '-',
    'ruleset': 'normal',
};

/*
 *  Initialization
 */

// Establish a WebSocket connection
var socket;
try {
    var webSocketURL;
    if (window.location.protocol == 'https:') {
        webSocketURL = 'wss://';
    } else {
        webSocketURL = 'ws://';
    }
    webSocketURL += window.location.host + '/ws';
    socket = new WebSocket(webSocketURL);
} catch(err) {
    showError('Failed to connect to the WebSocket server. Try logging out and back in again.');
}

/*
 *  WebSocket message event handlers
 */

// Create event handlers for all incoming messages
socket.onmessage = function(event) {
    // Log the incoming message
    var message = event.data;
    console.log('Recieved message:', message);

    // Parse the incoming message
    var m = message.match(/^(\w+) (.+)$/);
    var command;
    var data;
    if (m) {
        command = m[1];
        data = JSON.parse(m[2]);
    } else {
        console.error('Failed to parse message:', message);
    }

    /*
     *  Chat handlers
     */

    // roomList
    if (command === 'roomList') {
        // Keep track of who is in this channel
        roomList[data.room] = data.users;

    // roomHistory
    } else if (command === 'roomHistory') {
        for (var i = 0; i < data.history.length; i++) {
            if (data.room == 'global') {
                // Draw the global chat
                // TODO
            } else {
                // Draw the game chat
                // TODO
            }
        }

    /*
     *  Game handlers
     */

    // gameList
    } else if (command === 'gameList') {
        // Keep track of what games are currently going
        currentGames = data;

        // Populate the "Current games" area
        for (var i = 0; i < currentGames.length; i++) {
            // TODO
        }

    // gameCreated
    } else if (command === 'gameCreated') {
        // Keep track of what games are currently going
        currentGames.push(data)

        // Add the game to the "Current games" area
        // TODO

    /*
     *  Game action handlers
     */

    // actionPlay
    } else if (command === 'actionPlay') {
        // TODO

    // actionClue
    } else if (command === 'actionClue') {
        // TODO

    // actionDiscard
    } else if (command === 'actionDiscard') {
        // TODO

    // actionHands
    } else if (command === 'actionHands') {
        // TODO

    /*
     *  Miscellaneous handlers
     */

    // error
    } else if (command === 'error') {
        showError(data.msg);

    // Unknown command
    } else {
        showError('Unrecognized message: ' + message);
    }
}

/*
 *  UI buttons
 */

var rulesetChooseHTML = ' &nbsp;<span class="caret"></span>';
$('#ruleset-choose-normal').click(function(e) {
    createGameOptions.ruleset = 'normal';
    $('#ruleset-dropdown').html('Normal' + rulesetChooseHTML);
});
$('#ruleset-choose-black').click(function(e) {
    createGameOptions.ruleset = 'black';
    $('#ruleset-dropdown').html('Black' + rulesetChooseHTML);
});
$('#ruleset-choose-rainbow').click(function(e) {
    createGameOptions.ruleset = 'rainbow';
    $('#ruleset-dropdown').html('Rainbow' + rulesetChooseHTML);
});

$('#create-game-button').click(function(e) {
    createGameOptions.name = $('#game-name').val();
    var message = 'gameCreate ' + JSON.stringify(createGameOptions);
    socket.send(message);
    $('#create-game-modal').modal('toggle');
});

$('#past-games-button').click(function(e) {
    // TODO
});

/*
 *  Debug stuff
 */

$('#test-button').click(function(e) {
    showError('poop');
});

/*
 *   Miscellaneous functions
 */

function showError(message) {
    BootstrapDialog.show({
        title: 'Error',
        message: message,
        type: 'type-danger',
        buttons: [{
            label: 'Logout',
            action: function(dialog) {
                window.location.replace('/logout');
            },
        }],
    });
}
