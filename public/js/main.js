"use strict";

/*
 *  HANABI JAVASCRIPT - MAIN
 */

/*
TODO:
- get it working
- create game ruleset should remember from last time, ask from server
- random word generator
- create keyboard functionality for all ui stuff
*/

/*
 *  Global variables
 */

var username;
var currentGames = {};
var roomList = {};
var createGameOptions = {
    'name':    '-',
    'ruleset': 'normal',
};
var inPreGame = false; // Can be true or false
var gameID = false; // Can be false or the ID of the game
var gameState = {};

/*
 *  Initialization
 */

// Preload all of the card images
function preload(imageList) {
    for (var i = 0; i < imageList.length; i++) {
        var image = new Image();
        image.id = 'card-' + imageList[i];
        image.src = '/public/img/cards/' + imageList[i] + '.png';
        $('#card-images').append(image);
    }
}
preload([
    'b1', 'b2', 'b3', 'b4', 'b5',
    'g1', 'g2', 'g3', 'g4', 'g5',
    'y1', 'y2', 'y3', 'y4', 'y5',
    'r1', 'r2', 'r3', 'r4', 'r5',
    'p1', 'p2', 'p3', 'p4', 'p5',
    'k1', 'k2', 'k3', 'k4', 'k5',
    'm1', 'm2', 'm3', 'm4', 'm5',
    'back',
    'back-b', 'back-g', 'back-y', 'back-r', 'back-p',
    'back-k', 'back-m',
]);

// Dynamically resize the canvas to fill the browser window
var canvas = document.getElementById('hanabi-canvas');
var context = canvas.getContext('2d');
window.addEventListener('resize', hanabiResizeCanvas, false);

// Establish a WebSocket connection
var socket;
try {
    var webSocketURL;
    if (window.location.protocol === 'https:') {
        webSocketURL = 'wss://';
    } else {
        webSocketURL = 'ws://';
    }
    webSocketURL += window.location.host + '/ws';
    socket = new WebSocket(webSocketURL);
} catch(err) {
    errorShow('Failed to connect to the WebSocket server. Try logging out and back in again.');
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
        // Keep track locally of who is in this channel
        for (var i = 0; i < data.users.length; i++) {
            roomList[data.room] = {};
            roomList[data.room][data.users[i].name] = data.users[i];
        }

    // roomHistory
    } else if (command === 'roomHistory') {
        for (var i = 0; i < data.history.length; i++) {
            if (data.room === 'global') {
                // Draw the global chat
                // TODO
            } else {
                // Draw the game chat
                // TODO
            }
        }

    // roomJoined
    } else if (command === 'roomJoined') {
        // Keep track of who is in this channel
        roomList[data.room][data.user.name] = data.user;

        if (data.room === 'global') {
            // Add them to the global chat list
            // TODO
        } else {
            // Mark them as online inside the game
            // TODO
        }

    // roomLeft
    } else if (command === 'roomLeft') {
        // Remove them from the room list
        delete roomList[data.room][data.name];

        if (data.room === 'global') {
            // Remove them to the global chat list
            // TODO
        } else {
            // Mark them as offline inside the game
            // TODO
        }

    /*
     *  Game handlers
     */

    // gameList
    } else if (command === 'gameList') {
        // Keep track of what games are currently going
        for (var i = 0; i < data.length; i++) {
            currentGames[data[i].id] = data[i];
        }

        // Update the "Current games" area
        lobbyDrawCurrentGames();

        // Check to see if we are in any games
        var inAGame = false;
        for (var id in currentGames) {
            if (!currentGames.hasOwnProperty(id)) {
                continue;
            }

            for (var i = 0; i < currentGames[id].players.length; i++) {
                if (currentGames[id].players[i] === username) {
                    gameID = id;
                    if (currentGames[id].status === 'open') {
                        gameID = id;
                        lobbyDrawPregame();
                    } else if (currentGames[id].status === 'in progress') {
                        gameID = id;
                    } else {
                        errorShow('Failed to parse the status of game #' + id + ': ' + currentGames[id].status);
                    }
                    break;
                }
            }
        }

    // gameCreated
    } else if (command === 'gameCreated') {
        // Keep track of what games are currently going
        currentGames[data.id] = data;

        // Update the "Current games" area
        lobbyDrawCurrentGames();

        // Check to see if we created this game
        if (data.players[0] === username) { // There will only be one player in this game because it was just created
            gameID = data.id;
            lobbyDrawPregame();
        }

    // gameJoined
    } else if (command === 'gameJoined') {
        // Keep track of the people in each game
        currentGames[data.id].players.push(data.name);

        // Update the "Current games" area
        lobbyDrawCurrentGames();

        // Check to see if we joined this game
        if (data.name === username) {
            gameID = data.id;
        }

        // Check to see if we are in this game
        if (data.id === gameID) {
            lobbyDrawPregame();
        }

    // gameLeft
    } else if (command === 'gameLeft') {
        // Get the status of the game before we potentially delete it
        var currentStatus = currentGames[data.id].status;

        // Delete this person from the currentGames list
        if (currentGames[data.id].players.indexOf(data.name) !== -1) {
            currentGames[data.id].players.splice(currentGames[data.id].players.indexOf(data.name), 1)
        } else {
            errorShow('"' + data.name + '" left race #' + data.id + ', but they were not in the entrant list.');
            return;
        }

        // Check to see if this was the last person in the game, and if so, delete the game
        if (currentGames[data.id].players.length === 0) {
            delete currentGames[data.id];

        // Check to see if this person was the captain, and if so, make the next person in line the captain
        } else {
            if (currentGames[data.id].captain === data.name) {
                currentGames[data.id].captain = currentGames[data.id].players[0];
            }
        }

        // Update the "Current games" area
        lobbyDrawCurrentGames();

        // Check to see if we left this game
        if (data.name === username) {
            gameID = false;
            if (currentStatus === 'open') {
                lobbyLeavePregame();
            } else if (currentStatus === 'in progress') {
                showLobby();
            } else {
                errorShow('Failed to parse the status of game #' + data.id + ': ' + currentStatus);
            }

        // Check to see if someone else left a game that we are in
        } else if (data.id === gameID) {
            if (currentStatus === 'open') {
                lobbyDrawPregame();
            } else if (currentStatus === 'in progress') {
                showLobby();
            } else {
                errorShow('Failed to parse the status of game #' + data.id + ': ' + currentStatus);
            }
        }

    // gameSetStatus
    } else if (command === 'gameSetStatus') {
        // Update the status
        currentGames[data.id].status = data.status;

        // Check to see if we are in this game
        if (data.id === gameID) {
            if (currentGames[data.id].status === 'in progress') {
                // Do nothing; gameState will be given next and we will act on that
            } else if (currentGames[data.id].status === 'finished') {
                gameID = false;
            } else {
                errorShow('Failed to parse the status of game #' + data.id + ': ' + currentGames[data.id].status);
            }
        }

        // Remove the game if it is finished
        if (currentGames[data.id] === 'finished') {
            delete currentGames[data.id];
        }

    // gameState
    } else if (command === 'gameState') {
        // We started a new game or disconnected, so reset/initialize the variable that represents the game state
        gameState = data;

        // Leave the lobby and show the Hanabi GUI
        hanabiShow();

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

    /*
     *  Miscellaneous handlers
     */

    // username
    } else if (command === 'username') {
        username = data.name;

    // alert
    } else if (command === 'alert') {
        alertShow(data.msg);

    // error
    } else if (command === 'error') {
        errorShow(data.msg);

    // Unknown command
    } else {
        errorShow('Unrecognized message: ' + message);
    }
}

function socketSend(command, json) {
    var message = command + ' ' + JSON.stringify(json);
    socket.send(message);
    console.log('Sent message: ' + message);
}

/*
 *  Lobby UI functions
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
    socketSend('gameCreate', createGameOptions);
    $('#create-game-modal').modal('toggle');
});

$('#create-game-window').bind('keypress', function(e) {
    if (e.keyCode === 13) { // Enter
        $('#create-game-button').click();
    } else if (e.keyCode === 27) { // Esc
        $('#create-game-modal').modal('toggle');
    }
 });

$('#past-games-button').click(function(e) {
    alert('This isn\'t implemented yet.');
    // TODO
});

$('#leave-game').click(function(e) {
    var messageObject = {
        id: gameID
    }
    socketSend('gameLeave', messageObject);
});

/*
 *  Lobby functions
 */

function lobbyDrawCurrentGames() {
    // Populate the "Current games" area
    if (Object.keys(currentGames).length === 0) {
        $('#current-games').html('<li>No current games.</li>');
    } else {
        $('#current-games').html('');
        for (var id in currentGames) {
            if (!currentGames.hasOwnProperty(id)) {
                continue;
            }

            // Create the HTML for this game
            var insertedHTML = '<div class="row vertical-align"><div class="col-xs-2">';
            insertedHTML += '<button type="button" id="join-game-' + id + '" class="btn btn-sm btn-primary center-block">Join</button>';
            insertedHTML += '</div><div class="col-xs-10"><strong>Game #' + id;
            if (currentGames[id].name !== '-') {
                insertedHTML += ' &mdash; ' + currentGames[id].name;
            }
            insertedHTML += '<br />Players:</strong> ';
            for (var i = 0; i < currentGames[id].players.length; i++) {
                insertedHTML += currentGames[id].players[i] + ', ';
            }
            insertedHTML = insertedHTML.substring(0, insertedHTML.length - 2); // Chop off the trailing comma + space
            insertedHTML += '</div></div><br />'
            $('#current-games').append(insertedHTML);
            if (currentGames[id].status == 'open') {
                $('#join-game-' + id).unbind();
                $('#join-game-' + id).click(function() {
                    // Parse the ID
                    var m = $(this).attr('id').match(/join-game-(\d+)/);
                    var id;
                    if (m) {
                        id = m[1];
                    } else {
                        errorShow('Failed to parse the game ID from the button ID: ' + $(this).attr('id'));
                    }

                    // Start the game
                    id = parseInt(id);
                    var messageObject = {
                        id: id,
                    };
                    socketSend('gameJoin', messageObject);
                });
            } else if (currentGames[id].status == 'in progress') {
                $('#join-game-' + id).html('In Progress')
                $('#join-game-' + id).unbind();
                $('#join-game-' + id).fadeTo(0, 0.25);
                $('#join-game-' + id).css('cursor', 'default');
            } else {
                errorShow('Failed to parse the status of game #' + id + ':', currentGames[id].status);
            }
        }
    }
}

function lobbyDrawPregame() {
    // If we just joined this game, hide the "Current games" section and show the pregame section
    if (inPreGame === false) {
        inPreGame = true;
        $('#main-menu').fadeOut(400, function() {
            $('#pregame').fadeIn();
        });
    }

    // Game title
    var insertedHTML;
    if (currentGames[gameID].name === '-') {
        insertedHTML = gameID;
    } else {
        insertedHTML = gameID + ' &mdash; ' + currentGames[gameID].name;
    }
    $('#pregame-id').html(insertedHTML);

    // Game ruleset
    $('#pregame-ruleset').html(currentGames[gameID].ruleset.capitalizeFirstLetter())

    // Current players
    var insertedHTML = '';
    for (var i = 0; i < currentGames[gameID].players.length; i++) {
        // Find out if the player is a captain
        var playerIsCaptain = false;
        if (currentGames[gameID].captain === currentGames[gameID].players[i]) {
            playerIsCaptain = true;
        }

        // Insert the line for this player
        insertedHTML += '<li>'
        if (playerIsCaptain === true) {
            insertedHTML += '<strong>'
        }
        insertedHTML += currentGames[gameID].players[i]
        if (playerIsCaptain === true) {
            insertedHTML += '</strong> (captain)'
        }
        insertedHTML += '</li>';
    }
    $('#pregame-current-players').html(insertedHTML);

    // If we are a captain, make the start button clickable
    if (currentGames[gameID].captain === username) {
        $('#start-game').unbind();
        $('#start-game').click(function(e) {
            var messageObject = {
                id: gameID,
            };
            socketSend('gameStart', messageObject);
        });
        $('#start-game').fadeTo(400, 1);
        $('#start-game').css('cursor', 'pointer');

    // If we not a captain, grey out the start button
    } else {
        $('#start-game').unbind();
        $('#start-game').fadeTo(400, 0.25);
        $('#start-game').css('cursor', 'default');
    }
}

function lobbyLeavePregame() {
    inPreGame = false;
    $('#pregame').fadeOut(400, function() {
        $('#main-menu').fadeIn();
    });
}

function lobbyShow() {
    gameID = false;
    $('#hanabi').fadeOut(400, function() {
        $('#lobby').fadeIn();
    });
}

/*
 *  Hanabi functions
 */

function hanabiShow() {
    console.log('Started game #' + gameID + '.');

    // Show the Hanabi GUI
    $('#pregame').fadeOut();
    $('#main-menu').fadeIn();
    $('#lobby').fadeOut(400, function() {
        $('#hanabi').fadeIn();
    });

    // Draw the screen
    hanabiResizeCanvas();
    hanabiDraw();
}

function hanabiResizeCanvas() {
    // Based on the "size_stage" function from Keldon
    var windowWidth = window.innerWidth;
    var windowHeight = window.innerHeight;
    var canvasWidth, canvasHeight;

    if (windowWidth < 640) windowWidth = 640;
    if (windowHeight < 360) windowHeight = 360;

    var ratio = 1.777;

    if (windowWidth < windowHeight * ratio) {
        canvasWidth = windowWidth;
        canvasHeight = windowWidth / ratio;
    } else {
        canvasHeight = windowHeight;
        canvasWidth = windowHeight * ratio;
    }

    canvasWidth = Math.floor(canvasWidth);
    canvasHeight = Math.floor(canvasHeight);

    if (canvasWidth > 0.98 * windowWidth) canvasWidth = windowWidth;
    if (canvasHeight > 0.98 * windowHeight) canvasHeight = windowHeight;

    canvas.width = canvasWidth;
    canvas.height = canvasHeight;

    hanabiDraw();
}

function hanabiDraw() {
    var win_w = canvas.width;
    var win_h = canvas.height;
    var x, y, width, height, offset, radius;

    var suits = ['b', 'g', 'y', 'r', 'p'];
    if (gameState.ruleset == 'black') {
        suits.push('k')
    } else if (gameState.ruleset == 'rainbow') {
        suits.push('m')
    }

    // Draw the color card backs
    if (suits.length == 5) {
        y = .05;
        width = .075;
        height = .189;
         offset = 0;
        radius = .006;
    } else if (suits.length == 6) {
        y = .04;
        width = .06;
        height = .151;
        offset = .019;
        radius = .004;
    }
    for (var i = 0; i < suits.length; i++) {
        var image = document.getElementById('card-back-' + suits[i]);
        context.drawImage(
            image,
            (.183 + (width + .015) * i) * win_w,
            (.345 + offset) * win_h,
            width * win_w,
            height * win_h
         );
    }

    // Find out what player number we are
    var ourPlayerNum;
    for (var i = 0; i < gameState.hands.length; i++) {
        if (gameState.hands[i].name === username) {
            ourPlayerNum = gameState.hands[i].playerNum - 1;
            break;
        }
    }

    // Draw all the hands
    var hand_pos = {
        2: [
            { x: .19, y: .77, w: .42, h: .189, rot: 0 },
            { x: .19, y: .01, w: .42, h: .189, rot: 0 }
        ],
        3: [
            { x: .19, y: .77, w: .42, h: .189, rot: 0 },
            { x: .01, y: .71, w: .41, h: .189, rot: -78 },
            { x: .705, y: 0, w: .41, h: .189, rot: 78 }
        ],
        4: [
            { x: .23, y: .77, w: .34, h: .189, rot: 0 },
            { x: .015, y: .7, w: .34, h: .189, rot: -78 },
            { x: .23, y: .01, w: .34, h: .189, rot: 0 },
            { x: .715, y: .095, w: .34, h: .189, rot: 78 }
        ],
        5: [
            { x: .23, y: .77, w: .34, h: .189, rot: 0 },
            { x: .03, y: .77, w: .301, h: .18, rot: -90 },
            { x: .025, y: .009, w: .34, h: .189, rot: 0 },
            { x: .445, y: .009, w: .34, h: .189, rot: 0 },
            { x: .77, y: .22, w: .301, h: .18, rot: 90 }
        ]
    };
    var nump = gameState.hands.length;
    for (var i = 0; i < nump; i++) {
        var j = i - ourPlayerNum;

        if (j < 0) {
            j += nump;
        }
        console.log('j:', j);

        var image = document.getElementById('card-back');
        context.drawImage(
            image,
            hand_pos[nump][j].x * win_w,
            hand_pos[nump][j].y * win_h,
            hand_pos[nump][j].w * win_w,
            hand_pos[nump][j].h * win_h
        );

        /*
            rotationDeg: hand_pos[nump][j].rot,
            align: "center",
            reverse: j == 0
        */

    }

    // Discard area
    // TODO MAKE CLICKABLE
    /*context.rect(
        .8 * win_w,
        .6 * win_h,
        .2 * win_w,
        .4 * win_h
    );
    context.stroke();*/

    // Play area rectangle
    // TODO MAKE CLICKABLE
    /*context.rect(
        .183 * win_w,
        .3 * win_h,
        .435 * win_w,
        .189 * win_h
    );*/

    // TODO asdf

}

/*
 *   Miscellaneous functions
 */

function alertShow(message) {
    BootstrapDialog.show({
        title: 'Alert',
        message: message,
        type: 'type-warning',
        buttons: [{
            label: 'Ok',
            action: function(dialog) {
                dialog.close();
            },
        }],
    });
}

function errorShow(message) {
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

// Every time a modal is shown, if it has an autofocus element, focus on it
// From: http://stackoverflow.com/questions/14940423/autofocus-input-in-twitter-bootstrap-modal
$('.modal').on('shown.bs.modal', function() {
    $(this).find('[autofocus]').focus();
});

// Add the capitalizeFirstLetter function
// From: http://stackoverflow.com/questions/1026069/how-do-i-make-the-first-letter-of-a-string-uppercase-in-javascript?page=2&tab=active#tab-top
String.prototype.capitalizeFirstLetter = function() {
    return this.charAt(0).toUpperCase() + this.slice(1);
}
