package data

const TEMPLATE_FILES = `
{{ define "version_page" }}
	<html>
		<head>
			<meta charset="utf-8">
			<title>Current Weather Server</title>
		</head>
		<body>
	         <b>Current Weather Server</b>
             <br>
             <b>Version:</b> {{.}} 
		</body>
	</html>
{{ end }}

{{ define "get_longitude_latitude" }}
	<html>
		<head>
			<meta charset="utf-8">
			<title>Get Current Weather for...</title>
            <script>
                function displayError(msg) {
                    document.getElementById("validationError").innerHTML = msg;
                }

				function validateForm() {
                  displayError("");
				  let longitudeText = document.forms["longLatForm"]["longitude"].value.trim();
				  if (longitudeText == "") {
					displayError("Longitude missing");
					return false;
				  }

                  if (! /^[-+]?[0-9]*\.?[0-9]+$/.test(longitudeText)) {
                     displayError("Invalid number format for Longitude: " + longitudeText);
                     return false;
                  }

				  let latitudeText = document.forms["longLatForm"]["latitude"].value.trim();
				  if (latitudeText == "") {
					displayError("Latitude missing");
					return false;
				  }

                  if (! /^[-+]?[0-9]*\.?[0-9]+$/.test(latitudeText)) {
                     displayError("Invalid number format for Latitude: " + latitudeText );
                     return false;
                  }

                  longitude = parseFloat(longitudeText);


                  if (isNaN(longitude) || longitude < -180 || longitude > 180) {
                      displayError("Invalid longitude value.  Must be a number between -180 and 180");
                      return false;
                  }

                  latitude = parseFloat(latitudeText);

                  if (isNaN(latitude) || latitude < -90 || latitude > 90) {
                      displayError("Invalid latitude value.  Must be a number between -90 and 90");
                      return false;
                  }
                  console.log("Form valid");
                  return true;
				}
            </script>
		</head>
		<body>
	         <b>Current Weather Server</b>
             <br>
             <br>

			<form name="longLatForm" action="/displaycurrentweather.html" onSubmit="return validateForm()" method="get">
			  <label for="longitude">Longitude:</label>
			  <input type="text" id="longitude" name="longitude" value="0.0">
			  <label for="latitude">Latitude:</label>
			  <input type="text" id="latitude" name="latitude" value="0.0"> <br><br>

				<fieldset>
				  <legend>Temperature Unit:</legend>
				
				  <div>
					  <input type="radio" id="imperial" name="units" value="imperial" checked>
					  <label for="imperial">Fahrenheit</label>
				  
					  <input type="radio" id="metric" name="units" value="metric">
					  <label for="metric">Celsius</label>
				  
					  <input type="radio" id="standard" name="units" value="standard">
					  <label for="standard">Kelvin</label>
				  </div>
				</fieldset>

              <br><br>
			  <input type="submit" value="Get Current Weather">
			</form>

            <div id="validationError" style="color:red">
                 
            </div> 

		</body>
	</html>

{{ end }}

{{ define "display_current_weather" }}
	<html>
		<head>
			<meta charset="utf-8">
			<title>Current Weather Server</title>
		</head>
		<body>
	         <b>Current Weather at:</b>
             <br><br>
             <b>Latitude:</b> {{ .Lat }} <br>
             <b>Longitude:</b> {{ .Long }} <br>
             <br>
             <b>Data Collection Time:</b> {{ .DataCollectionTime }} <br>
             <b>Summary:</b> {{ .Summary }} <br>

             <br><br>

             <a href="getcurrentweather.html">Check another location</a> 
		</body>
	</html>
{{ end }}

{{ define "display_current_weather_error" }}
	<html>
		<head>
			<meta charset="utf-8">
			<title>Current Weather Server</title>
		</head>
		<body>
	         <b>Current Weather at:</b>
             <br><br>
             <b>Error:</b> {{ . }} <br>
             <br><br>

             <a href="getcurrentweather.html">Check another location</a> 
		</body>
	</html>
{{ end }}
`
