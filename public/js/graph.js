function draw_graph(e, dataset) {
	var w = 500;
	var h = 70;
	var padding = 20;

	var scaleY = d3.scale.linear()
		.domain([0, d3.max(dataset)])
		.range([0, h]);

	var scaleX = d3.scale.linear()
		.domain([48, 0]).range([padding/2, w-(padding/2)]);

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
		.attr("y", h - padding)
		.attr("width", w / dataset.length - 2)
		.attr("height", 0);

	svg.selectAll("rect")
		.transition()
		.delay(function(d, i) {
			return (dataset.length - i) * 20;
		})
		.duration(500)
		.attr("y", function(d) { return (h - padding)-scaleY(d); })
		.attr("height", function(d) {
			return scaleY(d);
		});
}
