package main

func template() {
	return
	`<!DOCTYPE html>
	<html>
	  <head>
	    <title>Writer deployments</title>
	  </head>
	  <body>
	    {{range .Customers}}
	    <H2>.CustomerName</H2>
	    <table>
	      <thead>
		<tr>
		  <th>Writer</th>
		  <th>Editor Service</th>
		  <th>Concept Backend</th>
		</tr>
	      </thead>
	      <tbody>
	      <tr>
		<th>Versions</th>
		<td>{{.WriterVersion}}</td>
		<td>{{.EditorServiceVersion}}</td>
		<td>{{.ConceptBackendVersion}}</td>
	      </tr>
	      <tr>
	        <th>Running count</th>
	        <td>{{.WriterCount}}</td>
	        <td>{{.EditorServiceCount}}</td>
	        <td>{{.ConceptBackendCount}}</td>
	      </tr>
	      </tbody>
	    </table>
	    {{end}}
	  </body>
	</html>`
}