# Anything Is PWA

This project allows you to create an installable PWA from any website

# How does it work ?

The index webpage will ask you for information about the PWA you want to create, and generate a custom manifest based on it. You will then be taken to a webpage that includes this manifest.

due to the PWA start URL needing to be on the same domain that installed the PWA, the start URL is `/redirect?url=<your website>`. This means you are dependent on this server even after installing the PWA. Feel free to open an issue if you find a workaround around this

# Licence

GPL-2-or-later
