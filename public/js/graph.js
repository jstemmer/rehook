var graphs = []

function draw_graph(e, dataset) {
	var w = parseInt(d3.select(e).style('width'), 10);
	var h = 70;
	var padding = 22;

	var scaleY = d3.scale.linear()
		.domain([0, d3.max(dataset)])
		.range([0, h-padding]);

	var scaleX = d3.scale.linear()
		.domain([48, 0]).range([padding, w-(padding*2)]);

	d3.select(e).select("svg").remove();
	var svg = d3.select(e)
		.append("svg")
		.attr("width", w)
		.attr("height", h);

	var axis = d3.svg
		.axis()
		.tickFormat(function(d){
			var current = new Date();
			current.setHours(current.getHours() - d);
			return current.getHours() + ":00";
		})
		.scale(scaleX);

	var axisY = d3.svg
		.axis()
		.scale(d3.scale.linear().domain([0, d3.max(dataset)]).range([(h-padding)-10, 0]))
		.ticks(2)
		.orient("left");

	var tip = d3.tip()
		.attr('class', 'd3-tip')
		.offset([-10,0])
		.html(function(d) { return d; });

	svg.call(tip);

	svg.append("g")
		.attr("class", "axis")
		.attr("transform", "translate(20," + (h-padding) + ")")
		.call(axis);

	svg.append("g")
		.attr("class", "axis")
		.attr("transform", "translate(" + ((padding-1)*2) + ", 10)")
		.call(axisY);

	svg.selectAll("rect")
		.data(dataset)
		.enter()
		.append("rect")
		.attr("x", function(d, i) {
			return w-scaleX(i);
		})
		.attr("y", h - padding - 2)
		.attr("width", w / dataset.length - 2)
		.attr("height", 0)
		.on('mouseover', tip.show)
		.on('mouseout', tip.hide);

	svg.selectAll("rect")
		.transition()
		.delay(function(d, i) {
			return (dataset.length - i) * 20;
		})
		.duration(500)
		.attr("y", function(d) { return (h - padding)-scaleY(d) - 2; })
		.attr("height", function(d) {
			return scaleY(d);
		});

}

function draw_graphs() {
	for (var i=0; i<graphs.length; i++) {
		draw_graph(graphs[i].e, graphs[i].dataset);
	}
}

d3.select(window).on('resize', draw_graphs);
