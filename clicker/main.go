package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

var (
	counter    float64
	autoClicks float64
	mutex      sync.Mutex
	errorMsg   string
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Counter with Boosts</title>
    <script>
		let autoClickerInterval;
		let currentCount = {{.Count}};
		let autoClicks = {{.AutoClicks}};
		let lastServerUpdate = Date.now();

		function updateCounter() {
			if (autoClicks > 0) {
				const now = Date.now();
				const deltaTime = (now - lastServerUpdate) / 1000;
				currentCount += autoClicks * deltaTime;
				document.getElementById('counterValue').textContent = currentCount.toFixed(2);
				lastServerUpdate = now;
			}
		}

		function updateServerCounter() {
			fetch('/update', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({count: currentCount})
			})
			.then(response => response.json())
			.then(data => {
				currentCount = parseFloat(data.count);
				document.getElementById('counterValue').textContent = currentCount.toFixed(2);
				lastServerUpdate = Date.now();
			})
			.catch(error => console.error('Error updating server:', error));
		}

		function startAutoClicker() {
			if (!autoClickerInterval) {
				autoClickerInterval = setInterval(() => {
					updateCounter();
					document.getElementById('counterValue').textContent = currentCount.toFixed(2);
				}, 50);
			}
		}

		function stopAutoClicker() {
			if (autoClickerInterval) {
				clearInterval(autoClickerInterval);
				autoClickerInterval = null;
			}
		}

		function increaseCounter() {
			fetch('/', { method: 'POST' })
				.then(response => response.json())
				.then(data => {
					currentCount = parseFloat(data.count);
					document.getElementById('counterValue').textContent = currentCount.toFixed(2);
					document.getElementById('errorMsg').textContent = data.error;
					autoClicks = parseFloat(data.autoClicks);
					updateAutoClicksDisplay();
					startAutoClicker();
					lastServerUpdate = Date.now();
				})
				.catch(error => {
					console.error('Error:', error);
					document.getElementById('errorMsg').textContent = "An error occurred. Please try again.";
				});
		}

		function buyBoost(cost, clicks) {
			fetch('/boost', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/x-www-form-urlencoded',
				},
				body: 'cost=' + cost + '&clicks=' + clicks
			})
			.then(response => response.json())
			.then(data => {
				if (data.success) {
					currentCount = parseFloat(data.count);
					document.getElementById('counterValue').textContent = currentCount.toFixed(2);
					autoClicks = parseFloat(data.autoClicks);
					updateAutoClicksDisplay();
					startAutoClicker();
				} else {
					currentCount = parseFloat(data.count);
				}
				document.getElementById('errorMsg').textContent = data.error;
				lastServerUpdate = Date.now();
			})
			.catch(error => {
				console.error('Error:', error);
				document.getElementById('errorMsg').textContent = "An error occurred. Please try again.";
			});
		}

		function updateAutoClicksDisplay() {
			document.getElementById('autoClicksSpeed').textContent = autoClicks.toFixed(2);
		}

		window.onload = function() {
			startAutoClicker();
			updateAutoClicksDisplay();
			setInterval(updateServerCounter, 500);
		};
    </script>
</head>
<body>
    <h1>Counter: <span id="counterValue">{{.Count}}</span></h1>
    <p>Auto-clicks per second: <span id="autoClicksSpeed">{{.AutoClicks}}</span></p>
    <p id="errorMsg" style="color: red;">{{.ErrorMsg}}</p>
    <button onclick="increaseCounter()">Increase (+1)</button>
    <button onclick="buyBoost(10, 1)">Buy +1 Autoclicker (Cost: 10)</button>
    <button onclick="buyBoost(50, 5)">Buy +5 Autoclicker (Cost: 50)</button>
    <button onclick="buyBoost(100, 10)">Buy +10 Autoclicker (Cost: 100)</button>
	<button onclick="buyBoost(1000, 100)">Buy +100 Autoclicker (Cost: 1000)</button>
	<button onclick="buyBoost(10000, 1000)">Buy +1000 Autoclicker (Cost: 10000)</button>
</body>
</html>
`

func main() {
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/boost", handleBoost)
	http.HandleFunc("/update", handleUpdate)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mutex.Lock()
		counter++
		count := counter
		errMsg := errorMsg
		clicks := autoClicks
		errorMsg = ""
		mutex.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":      count,
			"error":      errMsg,
			"autoClicks": clicks,
		})
		return
	}

	tmpl, err := template.New("counter").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mutex.Lock()
	count := counter
	errMsg := errorMsg
	clicks := autoClicks
	errorMsg = ""
	mutex.Unlock()

	data := struct {
		Count      float64
		ErrorMsg   string
		AutoClicks float64
	}{
		Count:      count,
		ErrorMsg:   errMsg,
		AutoClicks: clicks,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleBoost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cost, _ := strconv.ParseFloat(r.FormValue("cost"), 64)
	clicks, _ := strconv.ParseFloat(r.FormValue("clicks"), 64)

	mutex.Lock()
	defer mutex.Unlock()

	success := false
	if counter >= cost {
		counter -= cost
		autoClicks += clicks
		errorMsg = ""
		success = true
	} else {
		errorMsg = "Not enough points to buy this boost!"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":      counter,
		"error":      errorMsg,
		"autoClicks": autoClicks,
		"success":    success,
	})
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateData struct {
		Count float64 `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mutex.Lock()
	counter = updateData.Count
	mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":      counter,
		"autoClicks": autoClicks,
	})
}
