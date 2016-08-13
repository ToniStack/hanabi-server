// The Path.Rectangle constructor can take a Point and a Size object
/*var myBall = new Path.Circle(new Point(100, 100), 50);
myBall.fillColor = 'tomato';*/

var canvas = document.getElementById("hanabiCanvas");
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;


var suit_colors = [
	"#0044cc",
	"#00cc00",
	"#ccaa22",
	"#aa0000",
	"#6600cc",
	"#111111"
];


var cardWidth = 286;
var cardHeight = 406;
var x = cardWidth / 2;
var y = cardHeight / 2;



/*this.indicateRect = new Kinetic.Rect({
	x: 0,
	y: 0,
	width: config.width,
	height: config.height,
	cornerRadius: 6,
	strokeWidth: 12,
	stroke: "#ccccee",
	visible: false,
	listening: false
});*/


/*var point = new Point(0, 0);
var size = new Size(config.width, config.height);
var myRectangle = new Rectangle(point, size);
var cornerSize = new Size(20, 20);
var rec2 = new Path.RoundRectangle(myRectangle, cornerSize);
rec2.fillColor = 'tomato';*/

function drawCard() {
	var rectangle = new Rectangle(new Point(0, 0), new Point(x, y));
	var cornerSize = new Size(10, 10);
	var path = new Path.RoundRectangle(rectangle, cornerSize);
	path.fillColor = 'black';
}

//drawCard()



var VARIANT = constants.VARIANT;



function build_cards() {
	var cvs, ctx;
	var i, j, name;
	var xrad = cardWidth * 0.08, yrad = cardHeight * 0.08;
	var x, y;
	var rainbow = false, grad;
	var pathfuncs = [];

	if (this.variant == VARIANT.RAINBOW) rainbow = true;

	pathfuncs[0] = function() {
		ctx.beginPath();
		ctx.moveTo(75, 0);
		ctx.quadraticCurveTo(110, 60, 150, 100);
		ctx.quadraticCurveTo(110, 140, 75, 200);
		ctx.quadraticCurveTo(40, 140, 0, 100);
		ctx.quadraticCurveTo(40, 60, 75, 0);
	};

	pathfuncs[1] = function() {
		ctx.beginPath();
		ctx.moveTo(50, 180);
		ctx.lineTo(100, 180);
		ctx.quadraticCurveTo(80, 140, 75, 120);
		ctx.arc(110, 110, 35, 2.6779, 4.712, true);
		ctx.arc(75, 50, 35, 1, 2.1416, true);
		ctx.arc(40, 110, 35, 4.712, .4636, true);
		ctx.quadraticCurveTo(70, 140, 50, 180);
	};

	pathfuncs[2] = function() {
		var i;
		ctx.translate(75, 100);
		ctx.beginPath();
		ctx.moveTo(0, -75);
		for (i = 0; i < 5; i++)
		{
			ctx.rotate(Math.PI / 5);
			ctx.lineTo(0, -30);
			ctx.rotate(Math.PI / 5);
			ctx.lineTo(0, -75);
		}
	};

	pathfuncs[3] = function() {
		ctx.beginPath();
		ctx.moveTo(75, 65);
		ctx.bezierCurveTo(75, 57, 70, 45, 50, 45);
		ctx.bezierCurveTo(20, 45, 20, 82, 20, 82);
		ctx.bezierCurveTo(20, 100, 40, 122, 75, 155);
		ctx.bezierCurveTo(110, 122, 130, 100, 130, 82);
		ctx.bezierCurveTo(130, 82, 130, 45, 100, 45);
		ctx.bezierCurveTo(85, 45, 75, 57, 75, 65);
	};

	pathfuncs[4] = function() {
		ctx.beginPath();
		ctx.arc(75, 100, 75, 3, 4.3, true);
		ctx.arc(48, 83, 52, 5, 2.5, false);
	};

	pathfuncs[5] = function() {
		ctx.beginPath();
		ctx.beginPath();
		ctx.moveTo(50, 180);
		ctx.lineTo(100, 180);
		ctx.quadraticCurveTo(80, 140, 75, 120);
		ctx.arc(110, 110, 35, 2.6779, 5.712, true);
		ctx.lineTo(75, 0);
		ctx.arc(40, 110, 35, 3.712, .4636, true);
		ctx.quadraticCurveTo(70, 140, 50, 180);
	};

	if (rainbow)
	{
		pathfuncs[5] = function() {
			ctx.beginPath();
			ctx.moveTo(0, 140);
			ctx.arc(75, 140, 75, Math.PI, 0, false);
			ctx.lineTo(125, 140);
			ctx.arc(75, 140, 25, 0, Math.PI, true);
			ctx.lineTo(0, 140);
		};
	}

	var backpath = function(p) {
		ctx.beginPath();
		ctx.moveTo(p, yrad + p);
		ctx.lineTo(p, cardHeight - yrad - p);
		ctx.quadraticCurveTo(0, cardHeight, xrad + p, cardHeight - p);
		ctx.lineTo(cardWidth - xrad - p, cardHeight - p);
		ctx.quadraticCurveTo(cardWidth, cardHeight, cardWidth - p, cardHeight - yrad - p);
		ctx.lineTo(cardWidth - p, yrad + p);
		ctx.quadraticCurveTo(cardWidth, 0, cardWidth - xrad - p, p);
		ctx.lineTo(xrad + p, p);
		ctx.quadraticCurveTo(0, 0, p, yrad + p);
	};

	var drawshape = function() {
		ctx.shadowColor = "rgba(0, 0, 0, 0.9)";
		ctx.fill();
		ctx.shadowColor = "rgba(0, 0, 0, 0)";
		ctx.stroke();
	};

	for (i = 0; i < 6; i++)
	{
		for (j = 0; j <= 5; j++)
		{
			name = "card-" + i + "-" + j;

			cvs = document.createElement("canvas");
			cvs.width = cardWidth;
			cvs.height = cardHeight;

			ctx = cvs.getContext("2d");

			card_images[name] = cvs;

			backpath(4);

			ctx.fillStyle = "white";
			ctx.fill();

			ctx.save();
			ctx.clip();
			ctx.globalAlpha = 0.2;
			ctx.strokeStyle = "black";
			for (x = 0; x < cardWidth; x += 4 + Math.random() * 4)
			{
				ctx.beginPath();
				ctx.moveTo(x, 0);
				ctx.lineTo(x, cardHeight);
				ctx.stroke();
			}
			for (y = 0; y < cardHeight; y += 4 + Math.random() * 4)
			{
				ctx.beginPath();
				ctx.moveTo(0, y);
				ctx.lineTo(cardWidth, y);
				ctx.stroke();
			}
			ctx.restore();

			if (i == 5 && rainbow)
			{
				grad = ctx.createLinearGradient(0, 0, 0, cardHeight);

				grad.addColorStop(0, suit_colors[0]);
				grad.addColorStop(0.25, suit_colors[1]);
				grad.addColorStop(0.5, suit_colors[2]);
				grad.addColorStop(0.75, suit_colors[3]);
				grad.addColorStop(1, suit_colors[4]);

				ctx.fillStyle = grad;
				ctx.strokeStyle = grad;
			}
			else
			{
				ctx.fillStyle = suit_colors[i];
				ctx.strokeStyle = suit_colors[i];
			}

			backpath(4);

			ctx.save();
			ctx.globalAlpha = 0.3;
			ctx.fill();
			ctx.globalAlpha = 0.7;
			ctx.lineWidth = 8;
			ctx.stroke();
			ctx.restore();

			ctx.shadowBlur = 10;

			if (i == 5 && rainbow)
			{
				grad = ctx.createLinearGradient(0, 14, 0, 110);

				grad.addColorStop(0, suit_colors[0]);
				grad.addColorStop(0.25, suit_colors[1]);
				grad.addColorStop(0.5, suit_colors[2]);
				grad.addColorStop(0.75, suit_colors[3]);
				grad.addColorStop(1, suit_colors[4]);

				ctx.fillStyle = grad;
			}

			ctx.strokeStyle = "black";
			ctx.lineWidth = 2;
			ctx.lineJoin = "round";
			ctx.font = "bold 96pt Arial";
			ctx.shadowColor = "rgba(0, 0, 0, 0.9)";
			ctx.fillText(j.toString(), 19, 110);
			ctx.shadowColor = "rgba(0, 0, 0, 0)";
			ctx.strokeText(j.toString(), 19, 110);

			ctx.save();
			ctx.translate(cardWidth, cardHeight);
			ctx.rotate(Math.PI);
			ctx.shadowColor = "rgba(0, 0, 0, 0.9)";
			ctx.fillText(j.toString(), 19, 110);
			ctx.shadowColor = "rgba(0, 0, 0, 0)";
			ctx.strokeText(j.toString(), 19, 110);
			ctx.restore();

			if (i == 5 && rainbow)
			{
				grad = ctx.createRadialGradient(75, 150, 25, 75, 150, 75);

				grad.addColorStop(0, suit_colors[0]);
				grad.addColorStop(0.25, suit_colors[1]);
				grad.addColorStop(0.5, suit_colors[2]);
				grad.addColorStop(0.75, suit_colors[3]);
				grad.addColorStop(1, suit_colors[4]);

				ctx.fillStyle = grad;
			}

			ctx.lineWidth = 5;

			if (j == 1 || j == 3)
			{
				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.scale(0.4, 0.4);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();
			}

			if (j > 1)
			{
				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.translate(0, -120);
				ctx.scale(0.4, 0.4);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();

				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.translate(0, 120);
				ctx.scale(0.4, 0.4);
				ctx.rotate(Math.PI);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();
			}

			if (j > 3)
			{
				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.translate(-90, 0);
				ctx.scale(0.4, 0.4);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();

				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.translate(90, 0);
				ctx.scale(0.4, 0.4);
				ctx.rotate(Math.PI);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();
			}

			if (j == 0)
			{
				ctx.clearRect(0, 0, cardWidth, cardHeight);
			}

			if (j == 0 || j == 5)
			{
				ctx.save();
				ctx.translate(cardWidth / 2, cardHeight / 2);
				ctx.scale(0.6, 0.6);
				ctx.translate(-75, -100);
				pathfuncs[i]();
				drawshape();
				ctx.restore();
			}
		}
	}

	cvs = document.createElement("canvas");
	cvs.width = cardWidth;
	cvs.height = cardHeight;

	ctx = cvs.getContext("2d");

	card_images["card-back"] = cvs;

	backpath(4);

	ctx.fillStyle = "white";
	ctx.fill();

	ctx.save();
	ctx.clip();
	ctx.globalAlpha = 0.2;
	ctx.strokeStyle = "black";
	for (x = 0; x < cardWidth; x += 4 + Math.random() * 4)
	{
		ctx.beginPath();
		ctx.moveTo(x, 0);
		ctx.lineTo(x, cardHeight);
		ctx.stroke();
	}
	for (y = 0; y < cardHeight; y += 4 + Math.random() * 4)
	{
		ctx.beginPath();
		ctx.moveTo(0, y);
		ctx.lineTo(cardWidth, y);
		ctx.stroke();
	}
	ctx.restore();

	ctx.fillStyle = "black";

	backpath(4);

	ctx.save();
	ctx.globalAlpha = 0.5;
	ctx.fill();
	ctx.globalAlpha = 0.7;
	ctx.lineWidth = 8;
	ctx.stroke();
	ctx.restore();

	ctx.fillStyle = "#444444";
	ctx.lineWidth = 8;
	ctx.lineJoin = "round";

	ctx.translate(cardWidth / 2, cardHeight / 2);

	for (i = 0; i < 5; i++)
	{
		ctx.save();
		ctx.translate(0, -90);
		ctx.scale(0.4, 0.4);
		ctx.rotate(-i * Math.PI * 2 / 5);
		ctx.translate(-75, -100);
		pathfuncs[i]();
		drawshape();
		ctx.restore();

		ctx.rotate(Math.PI * 2 / 5);
	}
}


build_cards()



function build_ui() {
	var x, y, width, height, offset, radius;
	var i, j;
	var rect, img, text;
	var suits = 5;

	if (this.variant) suits = 6;

	var layers = stage.getLayers();

	for (i = 0; i < layers.length; i++)
	{
		layers[i].remove();
	}

	var background = new Kinetic.Image({
		x: 0,
		y: 0,
		width: win_w,
		height: win_h,
		image: ImageLoader.get("background")
	});

	bglayer.add(background);

	play_area = new Kinetic.Rect({
		x: .183 * win_w,
		y: .3 * win_h,
		width: .435 * win_w,
		height: .189 * win_h
	});

	discard_area = new Kinetic.Rect({
		x: .8 * win_w,
		y: .6 * win_h,
		width: .2 * win_w,
		height: .4 * win_h
	});

	no_discard_label = new Kinetic.Rect({
		x: .8 * win_w,
		y: .6 * win_h,
		width: .19 * win_w,
		height: .39 * win_h,
		stroke: "#df1c2d",
		strokeWidth: .007 * win_w,
		cornerRadius: .01 * win_w,
		visible: false
	});

	uilayer.add(no_discard_label);

	rect = new Kinetic.Rect({
		x: .8 * win_w,
		y: .6 * win_h,
		width: .19 * win_w,
		height: .39 * win_h,
		fill: "black",
		opacity: 0.2,
		cornerRadius: .01 * win_w
	});

	bglayer.add(rect);

	img = new Kinetic.Image({
		x: .82 * win_w,
		y: .62 * win_h,
		width: .15 * win_w,
		height: .35 * win_h,
		opacity: 0.2,
		image: ImageLoader.get("trashcan")
	});

	bglayer.add(img);

	rect = new Kinetic.Rect({
		x: .2 * win_w,
		y: (MHGA_show_more_log ? .235 : .24) * win_h,
		width: .4 * win_w,
		height: (MHGA_show_more_log ? .098 : .05) * win_h,
		fill: "black",
		opacity: 0.3,
		cornerRadius: .01 * win_h,
		listening: true
	});

	bglayer.add(rect);

	rect.on("click tap", function() {
		msgloggroup.show();
		overback.show();

		overlayer.draw();

		overback.on("click tap", function() {
			overback.off("click tap");

			msgloggroup.hide();
			overback.hide();

			overlayer.draw();
		});
	});

	message_prompt = new MultiFitText({
        align: "center",
        fontSize: .028 * win_h,
        fontFamily: "Verdana",
        fill: "#d8d5ef",
        shadowColor: "black",
        shadowBlur: 10,
        shadowOffset: { x: 0, y: 0 },
        shadowOpacity: 0.9,
        listening: false,
		x: .21 * win_w,
		y: ( MHGA_show_more_log ? .238 : .25) * win_h,
		width: .38 * win_w,
		height: (MHGA_show_more_log ? .095 : .03) * win_h,
		maxLines: (MHGA_show_more_log ? 3 : 1)
	});

	uilayer.add(message_prompt);

	overback = new Kinetic.Rect({
		x: 0,
		y: 0,
		width: win_w,
		height: win_h,
		opacity: 0.3,
		fill: "black",
		visible: false
	});

	overlayer.add(overback);

	msgloggroup = new HanabiMsgLog();

	overlayer.add(msgloggroup);

	rect = new Kinetic.Rect({
		x: .66 * win_w,
		y: .81 * win_h,
		width: .13 * win_w,
		height: .18 * win_h,
		fill: "black",
		opacity: 0.2,
		cornerRadius: .01 * win_w
	});

	bglayer.add(rect);

	for (i = 0; i < 3; i++)
	{
		rect = new Kinetic.Rect({
			x: (.67 + .04 * i) * win_w,
			y: .91 * win_h,
			width: .03 * win_w,
			height: .053 * win_h,
			fill: "black",
			opacity: 0.6,
			cornerRadius: .003 * win_w
		});

		bglayer.add(rect);
	}

	clue_label = new Kinetic.Text({
		x: .67 * win_w,
		y: .83 * win_h,
		width: .11 * win_w,
		height: .03 * win_h,
		fontSize: .03 * win_h,
		fontFamily: "Verdana",
		align: "center",
		text: "Clues: 8",
		fill: "#d8d5ef",
		shadowColor: "black",
		shadowBlur: 10,
		shadowOffset: { x: 0, y: 0 },
		shadowOpacity: 0.9
	});

	uilayer.add(clue_label);

	score_label = new Kinetic.Text({
		x: .67 * win_w,
		y: .87 * win_h,
		width: .11 * win_w,
		height: .03 * win_h,
		fontSize: .03 * win_h,
		fontFamily: "Verdana",
		align: "center",
		text: "Score: 0",
		fill: "#d8d5ef",
		shadowColor: "black",
		shadowBlur: 10,
		shadowOffset: { x: 0, y: 0 },
		shadowOpacity: 0.9
	});

	uilayer.add(score_label);

	rect = new Kinetic.Rect({
		x: .8 * win_w,
		y: .01 * win_h,
		width: .19 * win_w,
		height: .58 * win_h,
		fill: "black",
		opacity: 0.2,
		cornerRadius: .01 * win_w
	});

	bglayer.add(rect);

	clue_log = new HanabiClueLog({
		x: .81 * win_w,
		y: .02 * win_h,
		width: .17 * win_w,
		height: .56 * win_h
	});

	uilayer.add(clue_log);

	var pileback;

	y = .05;
	width = .075;
	height = .189;
	offset = 0;
	radius = .006;

	if (this.variant)
	{
		y = .04;
		width = .06;
		height = .151;
		offset = .019;
		radius = .004;
	}

	for (i = 0; i < suits; i++)
	{
		pileback = new Kinetic.Rect({
			fill: suit_colors[i],
			opacity: 0.4,
			x: (.183 + (width + .015) * i) * win_w,
			y: ((MHGA_show_more_log ? .345 : .3) + offset) * win_h,
			width: width * win_w,
			height: height * win_h,
			cornerRadius: radius * win_w
		});

		bglayer.add(pileback);

		pileback = new Kinetic.Image({
			x: (.183 + (width + .015) * i) * win_w,
			y: ((MHGA_show_more_log ? .345 : .3) + offset) * win_h,
			width: width * win_w,
			height: height * win_h,
			image: card_images["card-" + i + "-0"]
		});

		bglayer.add(pileback);

		pileback = new Kinetic.Rect({
			stroke: suit_colors[i],
			strokeWidth: 5,
			x: (.183 + (width + .015) * i) * win_w,
			y: ((MHGA_show_more_log ? .345 : .3) + offset) * win_h,
			width: width * win_w,
			height: height * win_h,
			cornerRadius: radius * win_w
		});

		bglayer.add(pileback);

		play_stacks[i] = new CardStack({
			x: (.183 + (width + .015) * i) * win_w,
			y: ((MHGA_show_more_log ? .345 : .3) + offset) * win_h,
			width: width * win_w,
			height: height * win_h
		});

		cardlayer.add(play_stacks[i]);

		discard_stacks[i] = new CardLayout({
			x: .81 * win_w,
			y: (.61 + y * i) * win_h,
			width: .17 * win_w,
			height: .17 * win_h
		});

		cardlayer.add(discard_stacks[i]);
	}

	rect = new Kinetic.Rect({
		x: .08 * win_w,
		y: .8 * win_h,
		width: .075 * win_w,
		height: .189 * win_h,
		fill: "black",
		opacity: 0.2,
		cornerRadius: .006 * win_w
	});

	bglayer.add(rect);

	drawdeck = new CardDeck({
		x: .08 * win_w,
		y: .8 * win_h,
		width: .075 * win_w,
		height: .189 * win_h,
		cardback: "card-back"
	});

	cardlayer.add(drawdeck);

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

	var shade_pos = {
		2: [
			{ x: .185, y: .762, w: .43, h: .205, rot: 0 },
			{ x: .185, y: .002, w: .43, h: .205, rot: 0 }
		],
		3: [
			{ x: .185, y: .762, w: .43, h: .205, rot: 0 },
			{ x: .005, y: .718, w: .42, h: .205, rot: -78 },
			{ x: .708, y: -.008, w: .42, h: .205, rot: 78 }
		],
		4: [
			{ x: .225, y: .762, w: .35, h: .205, rot: 0 },
			{ x: .01, y: .708, w: .35, h: .205, rot: -78 },
			{ x: .225, y: .002, w: .35, h: .205, rot: 0 },
			{ x: .718, y: .087, w: .35, h: .205, rot: 78 }
		],
		5: [
			{ x: .225, y: .762, w: .35, h: .205, rot: 0 },
			{ x: .026, y: .775, w: .311, h: .196, rot: -90 },
			{ x: .02, y: .001, w: .35, h: .205, rot: 0 },
			{ x: .44, y: .001, w: .35, h: .205, rot: 0 },
			{ x: .774, y: .215, w: .311, h: .196, rot: 90 }
		]
	};

	var name_pos = {
		2: [
			{ x: .18, y: .97, w: .44, h: .02 },
			{ x: .18, y: .21, w: .44, h: .02 }
		],
		3: [
			{ x: .18, y: .97, w: .44, h: .02 },
			{ x: .01, y: .765, w: .12, h: .02 },
			{ x: .67, y: .765, w: .12, h: .02 }
		],
		4: [
			{ x: .22, y: .97, w: .36, h: .02 },
			{ x: .01, y: .74, w: .13, h: .02 },
			{ x: .22, y: .21, w: .36, h: .02 },
			{ x: .66, y: .74, w: .13, h: .02 }
		],
		5: [
			{ x: .22, y: .97, w: .36, h: .02 },
			{ x: .025, y: .775, w: .116, h: .02 },
			{ x: .015, y: .199, w: .36, h: .02 },
			{ x: .435, y: .199, w: .36, h: .02 },
			{ x: .659, y: .775, w: .116, h: .02 }
		]
	};

	var nump = this.player_names.length;

	for (i = 0; i < nump; i++)
	{
		j = i - this.player_us;

		if (j < 0) j += nump;

		player_hands[i] = new CardLayout({
			x: hand_pos[nump][j].x * win_w,
			y: hand_pos[nump][j].y * win_h,
			width: hand_pos[nump][j].w * win_w,
			height: hand_pos[nump][j].h * win_h,
			rotationDeg: hand_pos[nump][j].rot,
			align: "center",
			reverse: j == 0
		});

		cardlayer.add(player_hands[i]);

		rect = new Kinetic.Rect({
			x: shade_pos[nump][j].x * win_w,
			y: shade_pos[nump][j].y * win_h,
			width: shade_pos[nump][j].w * win_w,
			height: shade_pos[nump][j].h * win_h,
			rotationDeg: shade_pos[nump][j].rot,
			cornerRadius: .01 * shade_pos[nump][j].w * win_w,
			opacity: 0.4,
			fillLinearGradientStartPoint: {x: 0, y: 0},
			fillLinearGradientEndPoint: {x: shade_pos[nump][j].w * win_w, y: 0},
			fillLinearGradientColorStops: [ 0, "rgba(0,0,0,0)", 0.9, "white" ]
		});

		if (j == 0)
		{
			rect.setFillLinearGradientColorStops([ 1, "rgba(0,0,0,0)", 0.1, "white" ]);
		}

		bglayer.add(rect);

		/*
		uilayer.add(new Kinetic.Rect({
			x: hand_pos[nump][j].x * win_w,
			y: hand_pos[nump][j].y * win_h,
			width: hand_pos[nump][j].w * win_w,
			height: hand_pos[nump][j].h * win_h,
			rotationDeg: hand_pos[nump][j].rot,
			stroke: "black",
			strokeWidth: 1
		}));
		*/

		name_frames[i] = new HanabiNameFrame({
			x: name_pos[nump][j].x * win_w,
			y: name_pos[nump][j].y * win_h,
			width: name_pos[nump][j].w * win_w,
			height: name_pos[nump][j].h * win_h,
			name: this.player_names[i]
		});

		uilayer.add(name_frames[i]);

		/*
		uilayer.add(new Kinetic.Rect({
			x: name_pos[nump][j].x * win_w,
			y: name_pos[nump][j].y * win_h,
			width: name_pos[nump][j].w * win_w,
			height: name_pos[nump][j].h * win_h,
			stroke: "black",
			strokeWidth: 1
		}));
		*/
	}


    no_clue_box = new Kinetic.Rect({
        x: .20 * win_w,
        y: .56 * win_h,
        width: .40 * win_w,
        height: .15 * win_h,
        cornerRadius: .01 * win_w,
        fill: "black",
        opacity: 0.5,
        cornerRadius: .01 * win_w,
        visible: false
    });

    uilayer.add(no_clue_box);

	no_clue_label = new Kinetic.Text({
		x: .15 * win_w,
		y: .585 * win_h,
		width: .5 * win_w,
		height: .19 * win_h,
		fontFamily: "Verdana",
		fontSize: .08 * win_h,
		strokeWidth: 1,
		text: "No Clues",
		align: "center",
		fill: "#df2c4d",
		stroke: "black",
		visible: false
	});

	uilayer.add(no_clue_label);


	clue_area = new Kinetic.Group({
		x: .15 * win_w,
		y: (MHGA_show_more_log ? .54 : .51) * win_h,
		width: .5 * win_w,
		height: .27 * win_h
	});

	clue_target_group = new ButtonGroup();
	clue_type_group = new ButtonGroup();

	var button;

	x = .21 * win_w - (nump - 2) * .044 * win_w;

	for (i = 0; i < nump - 1; i++)
	{
		j = (this.player_us + i + 1) % nump;

		button = new Button({
			x: x,
			y: 0,
			width: .08 * win_w,
			height: .025 * win_h,
			text: this.player_names[j],
			target_index: j
		});

		clue_area.add(button);

		x += .0875 * win_w;

		clue_target_group.add(button);
	}

	for (i = 1; i <= 5; i++)
	{
		button = new NumberButton({
			x: (.133 + (i - 1) * .049) * win_w,
			y: (MHGA_show_more_log ? .027 : .035) * win_h,
			width: .04 * win_w,
			height: .071 * win_h,
			number: i,
			clue_type: {type: CLUE.RANK, value: i}
		});

		clue_area.add(button);

		clue_type_group.add(button);
	}

	suits = 5;
	x = .133;

	if (this.variant == VARIANT.BLACKSUIT || this.variant == VARIANT.BLACKONE)
	{
		suits = 6;
		x = .108;
	}

	for (i = 0; i < suits; i++)
	{
		button = new ColorButton({
			x: (x + i * .049) * win_w,
			y: (MHGA_show_more_log ? 0.1 : .115) * win_h,
			width: .04 * win_w,
			height: .071 * win_h,
			color: suit_colors[i],
			clue_type: {type: CLUE.SUIT, value: i}
		});

		clue_area.add(button);

		clue_type_group.add(button);
	}

	submit_clue = new Button({
		x: .133 * win_w,
		y: (MHGA_show_more_log ? .172 : .195) * win_h,
		width: .236 * win_w,
		height: .051 * win_h,
		text: "Give Clue"
	});

	clue_area.add(submit_clue);

	clue_area.hide();

	uilayer.add(clue_area);

	rewind_replay = new Button({
		x: .15 * win_w,
		y: (MHGA_show_more_log ? .55 : .53) * win_h,
		width: .24 * win_w,
		height: .08 * win_h,
		text: "Rewind Replay",
		visible: false
	});

	uilayer.add(rewind_replay);

	advance_replay = new Button({
		x: .4 * win_w,
		y: (MHGA_show_more_log ? .55 : .53) * win_h,
		width: .24 * win_w,
		height: .08 * win_h,
		text: "Advance Replay",
		visible: false
	});

	uilayer.add(advance_replay);

	exit_game = new Button({
		x: .3 * win_w,
		y: (MHGA_show_more_log ? .65 : .63) * win_h,
		width: .2 * win_w,
		height: .1 * win_h,
		text: "Exit Game",
		visible: false
	});

	uilayer.add(exit_game);

	helpgroup = new Kinetic.Group({
		x: .1 * win_w,
		y: .1 * win_h,
		width: .8 * win_w,
		height: .8 * win_h,
		visible: false,
		listening: false
	});

	overlayer.add(helpgroup);

	rect = new Kinetic.Rect({
		x: 0,
		y: 0,
		width: .8 * win_w,
		height: .8 * win_h,
		opacity: 0.9,
		fill: "black",
		cornerRadius: .01 * win_w
	});

	helpgroup.add(rect);

	text = new Kinetic.Text({
		x: .03 * win_w,
		y: .03 * win_h,
		width: .74 * win_w,
		height: .74 * win_h,
		fontSize: .02 * win_w,
		fontFamily: "Verdana",
		fill: "white",
		text: "Welcome to Hanabi!\n\nWhen it is your turn, you may " +
		      "play a card by dragging it to the play stacks in the " +
		      "center of the screen.\n\nTo discard, drag the card " +
		      "to the discard area in the lower right.\n\nTo give " +
		      "a clue, select the player who will receive it, then " +
		      "select either the number or color of the clue you " +
		      "wish to give, then hit the Give Clue button.\n\n" +
		      "You may mouseover a card to see what clues have " +
		      "been given about it, or mouseover the clues in the " +
		      "log to see which cards it referenced."
	});

	helpgroup.add(text);

	help_button = new Button({
		x: .01 * win_w,
		y: .8 * win_h,
		width: .06 * win_w,
		height: .13 * win_h,
		text: "?"
	});

	uilayer.add(help_button);

	help_button.on("click tap", function() {
		helpgroup.show();
		overback.show();

		overlayer.draw();

		overback.on("click tap", function() {
			overback.off("click tap");

			helpgroup.hide();
			overback.hide();

			overlayer.draw();
		});
	});

	lobby_button = new Button({
		x: .01 * win_w,
		y: .94 * win_h,
		width: .06 * win_w,
		height: .05 * win_h,
		text: "Lobby"
	});

	uilayer.add(lobby_button);

	lobby_button.on("click tap", function() {
		lobby_button.off("click tap");
		ui.send_msg({type: "unattend_table", resp: {}});
		ui.lobby.game_ended();
	});

	if (ui.replay)
	{
		rewind_replay.show();
		advance_replay.show();
		exit_game.show();

		rewind_replay.on("click tap", function() {
			ui.perform_replay(-1);
		});

		advance_replay.on("click tap", function() {
			ui.perform_replay(1);
		});

		exit_game.on("click tap", function() {
			rewind_replay.off("click tap");
			advance_replay.off("click tap");
			exit_game.off("click tap");

			ui.send_msg({type: "abort", resp: {}});
			ui.lobby.game_ended();
		});
	}

	stage.add(bglayer);
	stage.add(uilayer);
	stage.add(cardlayer);
	stage.add(tiplayer);
	stage.add(overlayer);
};

//build_ui();
