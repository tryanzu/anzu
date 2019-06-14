banReason spam {
    effects = <<JS
        var factor = banN || 1;
        exports = {
            duration: 60 * factor,
            ip: true,
        }
    JS
}
banReason rude {}
banReason abuse {}
banReason spoofing {
    effects = <<JS
        var factor = banN || 1;
        exports = {
            duration: 60 * 24 * 7 * factor,
            ip: true,
        }
    JS
}
banReason other {}

// Flag reasons section.
flag spam {}
flag rude {}
flag duplicate {}
flag needs_review {}
flag other {}