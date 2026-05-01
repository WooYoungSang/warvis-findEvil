rule beacon_payload {
    meta:
        description = "Detects Cobalt Strike beacon payload markers"
        severity = "high"
        author = "FIND EVIL fixture"
    strings:
        $a = "Cobalt Strike"
        $b = "beacon_id="
    condition:
        any of them
}

rule rootkit_marker {
    meta:
        description = "Detects rootkit module loading"
        severity = "critical"
        author = "FIND EVIL fixture"
    strings:
        $r = "rootkit.ko"
    condition:
        $r
}
