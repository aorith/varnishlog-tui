vcl 4.0;

import std;

# I am the backend
backend default none;

sub vcl_recv {
    # ESI processing
    if (req.url ~ "^/esi/") {
        return (synth(700, "ESI 1"));
    }
    if (req.url == "/included-content") {
        return (synth(701, "ESI Content"));
    }

    if (req.url ~ "^/esi2/") {
        return (synth(702, "ESI 2"));
    }
    if (req.url == "/esi-content") {
        return (synth(703, "ESI Content 1"));
    }
    if (req.url ~ "/last-content") {
        return (synth(704, "ESI Content 2"));
    }

    set req.backend_hint = default;

    if (req.http.Test == "maybe" && req.url !~ "^/esi|(included|last)-content") {
        return (synth(500, "Oops"));
    }

    return (synth(200, "OK"));
}

sub vcl_deliver {
    if (req.http.xid ~ "0$") {
        set resp.http.X-Bye = "Adiós";
    } else if (req.http.xid ~ "1$") {
        set resp.http.X-Bye = "Au revoir";
    } else if (req.http.xid ~ "2$") {
        set resp.http.X-Bye = "Goodbye";
    } else if (req.http.xid ~ "3$") {
        set resp.http.X-Bye = "Tschüss";
    } else if (req.http.xid ~ "4$") {
        set resp.http.X-Bye = "Arrivederci";
    } else if (req.http.xid ~ "5$") {
        set resp.http.X-Bye = "Adeus";
    } else if (req.http.xid ~ "6$") {
        set resp.http.X-Bye = "До свидания";
    } else if (req.http.xid ~ "7$") {
        set resp.http.X-Bye = "さようなら";
    } else if (req.http.xid ~ "8$") {
        set resp.http.X-Bye = "안녕히 가세요";
    } else if (req.http.xid ~ "9$") {
        set resp.http.X-Bye = "再见";
    }
    std.log("X-Bye: " + resp.http.X-Bye);
}

sub vcl_synth {
    # Level 1 ESI
    if (resp.status == 700) {
        set resp.status = 200;
        set resp.http.Content-Type = "text/html; charset=utf-8";
        set resp.body = {"<html><body><p>Hello there</p><esi:include src='/included-content'/></body></html>"};
        return (deliver);
    }
    if (resp.status == 701) {
        set resp.status = 200;
        set resp.http.Content-Type = "text/html; charset=utf-8";
        set resp.body = {"<p>This is included content via ESI.</p>"};
        return (deliver);
    }

    # Level 1 and 2 ESI
    if (resp.status == 702) {
        set resp.status = 200;
        set resp.http.Content-Type = "text/html; charset=utf-8";
        set resp.body = {"<html><body><p>Main</p><esi:include src='/esi-content'/></body></html>"};
        return (deliver);
    }
    if (resp.status == 703) {
        set resp.status = 200;
        set resp.http.Content-Type = "text/html; charset=utf-8";
        set resp.body = {"<p>Level 1 ESI.</p><esi:include src='/last-content?a=1'/> <hr> <esi:include src='/last-content?a=2'/>"};
        return (deliver);
    }
    if (resp.status == 704) {
        set resp.status = 200;
        set resp.http.Content-Type = "text/html; charset=utf-8";
        set resp.body = {"<p>Level 2 ESI.</p>"};
        return (deliver);
    }
}
