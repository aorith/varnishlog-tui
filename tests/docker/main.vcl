vcl 4.0;

import std;

backend varnishb {
    .host = "varnishb";
    .port = "80";
}

sub vcl_recv {
    # Random headers
    set req.http.xid = req.xid;
    if (req.http.xid ~ "[012]$") {
        set req.http.X-Test-Header = "Test Value";
    }
    if (req.http.xid ~ "[345]$") {
        set req.http.Another-Header = "Hi there";
    }
    if (req.http.xid ~ "[678]$") {
        set req.http.Legit-Request = "1";
        set req.http.Test = "maybe";
    }
    if (req.http.xid ~ "9$") {
        set req.http.Authorization = "goahead";
    }

    if (req.http.host ~ "^www\.example[12]\.com$") {
        set req.backend_hint = varnishb;
    } else {
        return (synth(404, "Host Not Found"));
    }
}

sub vcl_backend_response {
    if (bereq.url ~ "^/esi") {
        set beresp.do_esi = true;
    }

    if (bereq.url ~ "last") {
        set beresp.http.Cache-Control = "max-age=10 s-maxage=15";
        set beresp.ttl = 15s;
    }
}

sub vcl_deliver {
    # This Varnish instance can speak multiple languages! What a polyglot genius 😄
    if (req.http.xid ~ "0$") {
        set resp.http.X-Greet = "Hola";
    } else if (req.http.xid ~ "1$") {
        set resp.http.X-Greet = "Bonjour";
    } else if (req.http.xid ~ "2$") {
        set resp.http.X-Greet = "Hello";
    } else if (req.http.xid ~ "3$") {
        set resp.http.X-Greet = "Hallo";
    } else if (req.http.xid ~ "4$") {
        set resp.http.X-Greet = "Ciao";
    } else if (req.http.xid ~ "5$") {
        set resp.http.X-Greet = "Olá";
    } else if (req.http.xid ~ "6$") {
        set resp.http.X-Greet = "Привет";
    } else if (req.http.xid ~ "7$") {
        set resp.http.X-Greet = "こんにちは";
    } else if (req.http.xid ~ "8$") {
        set resp.http.X-Greet = "안녕하세요";
    } else if (req.http.xid ~ "9$") {
        set resp.http.X-Greet = "您好";
    }
    std.log("X-Greet: " + resp.http.X-Greet);
}
