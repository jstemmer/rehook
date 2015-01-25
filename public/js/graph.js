var graphs = []

function draw_graph(e, dataset) {
	var w = parseInt(d3.select(e).style('width'), 10);
	var h = 70;
	var padding = 22;

	var scaleY = d3.scale.linear()
		.domain([0, d3.max(dataset)])
		.range([0, h]);

	var scaleX = d3.scale.linear()
		.domain([48, 0]).range([padding/2, w-(padding/2)]);

	d3.select(e).select("svg").remove();
	var svg = d3.select(e)
		.append("svg")
		.attr("width", w)
		.attr("height", h);

	var axis = d3.svg
		.axis()
		.scale(scaleX);

	svg.append("g")
		.attr("class", "axis")
		.attr("transform", "translate(0," + (h-padding) + ")")
		.call(axis);

	svg.selectAll("rect")
		.data(dataset)
		.enter()
		.append("rect")
		.attr("x", function(d, i) {
			return w-scaleX(i);
		})
		.attr("y", h - padding - 2)
		.attr("width", w / dataset.length - 2)
		.attr("height", 0);

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
