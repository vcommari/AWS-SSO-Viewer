function sortTable(n) {
    var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
    table = document.getElementById("myTable");
    switching = true;
    // Set the sorting direction to ascending:
    dir = "asc";
    /* Make a loop that will continue until
    no switching has been done: */
    while (switching) {
        // Start by saying: no switching is done:
        switching = false;
        rows = table.rows;
        /* Loop through all table rows (except the
        first, which contains table headers): */
        console.log("here")
        for (i = 1; i < (rows.length - 1); i++) {
            // Start by saying there should be no switching:
            shouldSwitch = false;
            /* Get the two elements you want to compare,
            one from current row and one from the next: */
            x = rows[i].getElementsByTagName("TD")[n];
            y = rows[i + 1].getElementsByTagName("TD")[n];
            //console.log(x)
            /* Check if the two rows should switch place,
            based on the direction, asc or desc: */
            if (dir == "asc") {
                if (x.innerHTML.toLowerCase() > y.innerHTML.toLowerCase()) {
                    // If so, mark as a switch and break the loop:
                    shouldSwitch = true;
                    break; function myFunction() {
                        // Declare variables
                        var input, filter, table, tr, td, i, txtValue;
                        input = document.getElementById("myInput");
                        filter = input.value.toUpperCase();
                        table = document.getElementById("myTable");
                        tr = table.getElementsByTagName("tr");

                        // Loop through all table rows, and hide those who don't match the search query
                        for (i = 0; i < tr.length; i++) {
                            td = tr[i].getElementsByTagName("td")[0];
                            if (td) {
                                txtValue = td.textContent || td.innerText;
                                if (txtValue.toUpperCase().indexOf(filter) > -1) {
                                    tr[i].style.display = "";
                                } else {
                                    tr[i].style.display = "none";
                                }
                            }
                        }
                    }
                }
            } else if (dir == "desc") {
                if (x.innerHTML.toLowerCase() < y.innerHTML.toLowerCase()) {
                    // If so, mark as a switch and break the loop:
                    shouldSwitch = true;
                    break;
                }
            }
        }
        if (shouldSwitch) {
            /* If a switch has been marked, make the switch
            and mark that a switch has been done: */
            rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
            switching = true;
            // Each time a switch is done, increase this count by 1:
            switchcount++;
        } else {
            /* If no switching has been done AND the direction is "asc",
            set the direction to "desc" and run the while loop again. */
            if (switchcount == 0 && dir == "asc") {
                dir = "desc";
                switching = true;
            }
        }
    }
}


function search() {
    // Declare variables
    var input, filter, table, tr, td, i, txtValue;
    input = document.getElementById("myInput");
    filter = input.value.toUpperCase();
    table = document.getElementById("myTable");
    tr = table.getElementsByTagName("tr");
    th = table.getElementsByTagName("th");
    // Loop through all table rows, and hide those who don't match the search query
    for (i = 1; i < tr.length; i++) {
        tr[i].style.display = "none";
        for (var j = 0; j < th.length; j++) {
            td = tr[i].getElementsByTagName("td")[j];
            if (td) {
                txtValue = td.textContent || td.innerText;
                if (txtValue.toUpperCase().indexOf(filter) > -1) {
                    tr[i].style.display = "";
                    break;
                }
            }
        }
    }
}

//this function is in the event listener and will execute on page load
function populate_table() {
    // Populate colunm titles
    var table = document.getElementById('myTable');
    var tr = document.createElement('tr');
    var page = getParameterByName('page');
    console.log(page);
    switch(page) {
        case "ps":
            tr.innerHTML =
                '<th onclick="sortTable(0)">Name</th>' +
                '<th onclick="sortTable(1)">Arn</th>';
            var inlinetable = document.getElementById('inlinetable');
            var inlinetr = document.createElement('tr');
            inlinetr.innerHTML = '<th>Inline Policy</th>';
            inlinetable.appendChild(inlinetr);
            break;
        case "accounts":
            tr.innerHTML =
                '<th onclick="sortTable(0)">Group</th>' +
                '<th onclick="sortTable(1)">PermissionSet</th>';
            break;
        default:
            tr.innerHTML =
                '<th onclick="sortTable(0)">Account name</th>' +
                '<th onclick="sortTable(1)">Account Id</th>';
                break;            
    }
    table.appendChild(tr);
    console.log(window.location.host)

    // Fetch and display data
    switch(page) {
        case "ps":
            var arn = getParameterByName('arn');
            var json_url = 'http://' + window.location.host + '/getpspolicies?arn=' + arn;
            break;
        case "accounts":
            var account = getParameterByName('account');
            var json_url = 'http://' + window.location.host + '/getaccount/' + account;
            break;
        default:
            var json_url = 'http://' + window.location.host + '/accountslist';
    }
    var xhr = new XMLHttpRequest();
    xhr.open('GET', json_url, true);

    xhr.onload = function () {
        data = JSON.parse(xhr.response)
        switch(page) {
            case "ps":
                for (let k in data) {
                    var tr = document.createElement('tr');
                    tr.innerHTML =
                        '<td>' + k + '</td>' +
                        '<td>' + data[k] + '</td>';
                    table.appendChild(tr);
                };
                // Inline policy
                json_url2 = 'http://' + window.location.host + '/getpsinline?arn=' + arn;
                var xhr2 = new XMLHttpRequest();
                xhr2.open('GET', json_url2, true);
                xhr2.onload = function () {
                    var result = xhr2.response;
                    prettyResult = JSON.stringify(JSON.parse(result), null, 2);
                    var inlinetable = document.getElementById('inlinetable');
                    var inlinetr = document.createElement('tr');
                    inlinetr.innerHTML = "<pre>" + prettyResult + "</pre>";
                    inlinetable.appendChild(inlinetr); 
                }
                xhr2.send(null);
                break;
            case "accounts":
                for (let k in data) {
                    var tr = document.createElement('tr');
                    tr.innerHTML =
                        '<td>' + k + '</td>' +
                        '<td>' + data[k] + '</td>';
                    table.appendChild(tr);
                }
                break;
            default:
                data.forEach(function (object) {
                    var tr = document.createElement('tr');
                    tr.innerHTML =
                        '<td>' + object.Name + '</td>' +
                        '<td>' + object.Id + '</td>';
                    table.appendChild(tr);
                });
        }
        //append_json(accounts)
    };
    xhr.send(null);
}

//this function appends the json data to the table 'myTable'
function append_json(data) {
    var table = document.getElementById('myTable');
    data.forEach(function (object) {
        var tr = document.createElement('tr');
        tr.innerHTML =
            '<td>' + object.Name + '</td>' +
            '<td>' + object.Id + '</td>';
        table.appendChild(tr);
    });
}


//account JS
function getParameterByName(name, url = window.location.href) {
    name = name.replace(/[\[\]]/g, '\\$&');
    var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, ' '));
}

function get_json_account() {
    console.log(window.location.pathname)
    var account = getParameterByName('account');
    var json_url = 'http://localhost:8080/getaccount/' + account;
    var xhr = new XMLHttpRequest();
    xhr.open('GET', json_url, true);

    xhr.onload = function () {
        accounts = JSON.parse(xhr.response)
        append_account_json(accounts)
    };
    xhr.send(null);
}
//this function appends the json data to the table 'myTable'
function append_account_json(data) {
    var table = document.getElementById('myTable');
    for (let k in data) {
        var tr = document.createElement('tr');
        tr.innerHTML =
            '<td>' + k + '</td>' +
            '<td>' + data[k] + '</td>';
        table.appendChild(tr);
    };
}

function get_json_ps() {
    var arn = getParameterByName('arn');
    var json_url = 'http://localhost:8080/getpspolicies?arn=' + arn;
    var xhr = new XMLHttpRequest();
    xhr.open('GET', json_url, true);

    xhr.onload = function () {
        policy = JSON.parse(xhr.response)
        append_ps_json(policy)
    };
    xhr.send(null);

    // Inline policy
    json_url2 = 'http://localhost:8080/getpsinline?arn=' + arn;
    var xhr2 = new XMLHttpRequest();
    xhr2.open('GET', json_url2, true);

    xhr2.onload = function () {
        var result = xhr2.response;
        prettyResult = JSON.stringify(JSON.parse(result), null, 2);
        var inlinetable = document.getElementById('inlinetable');
        var inlinetr = document.createElement('tr');
        inlinetr.innerHTML = "<pre>" + prettyResult + "</pre>";
        inlinetable.appendChild(inlinetr);
        var test = document.getElementById('test');
    };
    
    xhr2.send(null);
}
//this function appends the json data to the table 'myTable'
function append_ps_json(data) {
    var table = document.getElementById('myTable');
    for (let k in data) {
        var tr = document.createElement('tr');
        tr.innerHTML =
            '<td>' + k + '</td>' +
            '<td>' + data[k] + '</td>';
        table.appendChild(tr);
    };
}
