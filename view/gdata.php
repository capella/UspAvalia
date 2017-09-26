<style>
.disparity-line {
   pointer-events: none;
   stroke: #37241D;
   stroke-width: 2;
   opacity: 0.65;
   shape-rendering: crispEdges;
}

.reference-line {
   stroke: #CCCCCC;
   stroke-width: 1;
   shape-rendering: crispEdges;
}

.axis {
   font: 10px sans-serif;
}

.axis path,
.axis line {
   fill: none;
   stroke: #000;
   shape-rendering: crispEdges;
}

.grid text {
   display: none;
}

.grid line {
   stroke: #E9E9E9;
   stroke-dasharray: 2,5;
}
.grid path{
   fill: none;
}

</style>
<script src="https://cdnjs.cloudflare.com/ajax/libs/d3/3.5.17/d3.min.js"></script>
<script>

$( ".graph" ).each(function() {
  var max = 10;
  var min = 0;
  var media = parseInt($(this).attr('avg'));
  var std = parseInt($(this).attr('std'));

  var margin = {top: 0, right: 10, bottom: 10, left: 15},
    width = $(this).width() - margin.left - margin.right,
    height = 50 - margin.top - margin.bottom;
  
  var svgContainer = d3.select($(this).get(0)).append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.right + margin.bottom)
  .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  
  //Draw the line
  var linah = 7.0/10;

  var constante = width/(max-min);
  svgContainer.append("line")
     .attr("x1", constante*media)
     .attr("y1", height*linah)
     .attr("x2", constante*Math.max(media-std,min))
     .attr("y2", height*linah)
     .classed("disparity-line", !0);

  svgContainer.append("line")
     .attr("x1", constante*Math.min(media+std,max))
     .attr("y1", height*linah)
     .attr("x2", constante*media)
     .attr("y2", height*linah)
     .classed("disparity-line", !0);

              
    var x = d3.scale.linear()
    .domain([min, max])
    .range([0, width]);

    var xAxis = d3.svg.axis()
        .scale(x)
        .orient("bottom")
        .ticks(10, '')
        .tickSize(6, 0);
    
    var xAxis2 = d3.svg.axis()
        .scale(x)
        .ticks(10)
        .orient("bottom")
        .tickSize(-height);
        
     svgContainer.append("g")
    .attr("class", "grid")
     .attr("transform", "translate(0," + height + ")")
    .call(xAxis2);  
    
    
        
    svgContainer.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + height + ")")
    .call(xAxis);
    
    var pointg = svgContainer.append("g").classed("mean-point", !0),
    ponto = pointg.append("path")
      .attr("d", "M-4.5,0L0,-4.5L4.5,0L0,4.5Z");
    pointg.attr("transform", "translate(" + (constante*media) + ","+height*linah+")");
});
</script>