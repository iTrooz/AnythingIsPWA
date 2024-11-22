let manualName = false;
let manualImage = false;

let form = document.forms["form"];

form["name"].addEventListener("input", function() {
    manualName = !!this.value.trim();
});

form["icon_url"].addEventListener("input", function() {
    manualImage = !!this.value.trim();
});

document.getElementById("start_url").addEventListener("input", function() {
    if (manualName&&manualImage) return;

    fetch(`/getWebsiteInfos?url=${encodeURIComponent(this.value)}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (!manualName) {
            form["name"].value = data.title || "";
        }
        if (!manualImage) {
            form["icon_url"].value = data.icon_url || "";
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
});

function isURL(s) {
    try {
        new URL(s);
        return true;
    } catch (e) {
        return false;
    }
}

function validateIcon(icon_url) {
    return new Promise(resolve=> {
        let iconUrl = form["icon_url"].value;
        let img = new Image();
        img.onload = function() {
            let ext = iconUrl.split('.').pop();
            if (img.width >= 144 && img.height >= 144 && ext == "png") {
                resolve(true);
            } else {
                alert(`Icon must be a PNG image with a size of at least 144x144 pixels (Currently width=${img.width}, height=${img.height}, type=${ext})`);
                resolve(false);
            }
        };
        img.onerror = function(e) {
            alert("Failed to load the icon image");
            resolve(false);
        };
        img.src = iconUrl;
    })
}

async function validateForm() {
    let form = document.forms["form"];

    if (!form["name"].value.trim()) {
        alert("Name must be filled out");
        return;
    }else if (!isURL(form["start_url"].value)) {
        alert("URL of the website is not valid");
        return;
    } else if (!isURL(form["icon_url"].value)) {
        alert("URL of the icon is not valid");
        return;
    }

    hint.innerHTML = "Checking icon validity..";
    if (!await validateIcon(form["icon_url"].value)) {
        hint.innerHTML = "";
        return false;
    }
    hint.innerHTML = "";

    form.submit();
} 