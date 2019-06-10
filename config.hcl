banReason spam {
    effects = <<JS
        var factor = banN || 1;
        exports = {
            duration: 60 * factor,
            ip: true,
        }
    JS
}

banReason spoofing {
    effects = <<JS
        var factor = banN || 1;
        exports = {
            duration: 60 * 24 * 7 * factor,
            ip: true,
        }
    JS
}