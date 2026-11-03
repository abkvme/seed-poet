#!/usr/bin/env python3
"""
Create phonetic rhyme groups for BIP39 words.
Groups words by their actual pronunciation endings, not spelling.
"""

import os

import datafile

# Phonetic rhyme groups - words that actually rhyme when spoken
# Format: group_name -> list of BIP39 words that rhyme

PHONETIC_RHYME_GROUPS = {
    # -AY sound (day, way, say)
    "ay": ["away", "betray", "clay", "day", "decay", "delay", "display", "essay",
           "gray", "hay", "holiday", "may", "okay", "pay", "play", "pray", "ray",
           "say", "spray", "stay", "stray", "subway", "survey", "sway", "today",
           "tray", "way"],

    # -EE sound (see, free, key)
    "ee": ["agree", "bee", "degree", "fee", "flee", "free", "key", "knee", "sea",
           "see", "spree", "three", "tree"],

    # -AIN sound (rain, brain, train)
    "ain": ["again", "brain", "chain", "domain", "drain", "explain", "gain", "grain",
            "main", "obtain", "pain", "plain", "rain", "remain", "strain", "sustain",
            "train", "vain"],

    # -ING sound (ring, bring, thing)
    "ing": ["bring", "cling", "fling", "king", "ring", "sing", "spring", "sting",
            "string", "swing", "thing", "wing"],

    # -IGHT sound (light, night, right)
    "ight": ["bright", "delight", "flight", "height", "knight", "light", "might",
             "night", "right", "sight", "slight", "tight"],

    # -OW (long O) sound (know, grow, show) - single syllable or stress on last syllable
    "oh": ["ago", "below", "blow", "flow", "glow", "grow", "know", "low", "owe",
           "row", "show", "slow", "snow", "throw"],

    # -OW unstressed ending (borrow, follow, shadow)
    "ow_unstressed": ["borrow", "elbow", "follow", "hollow", "meadow", "narrow",
                      "pillow", "shadow", "shallow", "sorrow", "tomorrow", "widow",
                      "window", "yellow"],

    # -OON sound (moon, soon, balloon)
    "oon": ["afternoon", "balloon", "cartoon", "lagoon", "moon", "noon", "raccoon",
            "soon", "spoon", "tune"],

    # -ACE sound (face, place, space)
    "ace": ["base", "case", "chase", "embrace", "erase", "face", "grace", "place",
            "race", "replace", "space", "trace"],

    # -ATE sound (late, create, state)
    # Note: removed "climate" (KLY-mit), "private" (PRY-vit), "graduate" (noun: GRAD-yoo-it)
    "ate": ["abbreviate", "celebrate", "create", "date", "debate", "educate",
            "estate", "fate", "gate", "generate", "great", "hate", "late",
            "locate", "mandate", "migrate", "plate", "rate", "rotate",
            "situate", "skate", "state", "straight", "translate", "update", "wait",
            "weight"],

    # -EAM sound (dream, team, stream)
    "eam": ["cream", "dream", "extreme", "gleam", "ream", "scheme", "scream", "seam",
            "steam", "stream", "team", "theme"],

    # -EAR sound (clear, fear, near)
    "ear": ["appear", "career", "clear", "dear", "disappear", "ear", "fear", "gear",
            "hear", "near", "pioneer", "severe", "shear", "sincere", "sphere", "steer",
            "year"],

    # -OOD sound (good, wood, should)
    "ood": ["childhood", "could", "good", "hood", "neighborhood", "should", "stood",
            "wood", "would"],

    # -UCK sound (duck, luck, truck)
    "uck": ["chuck", "duck", "luck", "muck", "pluck", "struck", "stuck", "truck", "tuck"],

    # -UNK sound (chunk, drunk,unk)
    "unk": ["bunk", "chunk", "drunk", "funk", "junk", "monk", "punk", "shrunk",
            "skunk", "slunk", "spunk", "stunk", "sunk", "trunk"],

    # -ASH sound (cash, crash, flash)
    "ash": ["ash", "brash", "cash", "clash", "crash", "dash", "flash", "gash", "hash",
            "lash", "mash", "rash", "slash", "smash", "splash", "stash", "trash"],

    # -EST sound (best, test, west)
    "est": ["arrest", "best", "breast", "chest", "contest", "crest", "digest", "gest",
            "guest", "harvest", "honest", "invest", "jest", "manifest", "nest", "pest",
            "protest", "quest", "request", "rest", "suggest", "test", "vest", "west",
            "zest"],

    # -ICK sound (kick, quick, trick)
    # Note: removed "magic", "panic", "topic" - they end in "-ic" not "-ick"
    "ick": ["brick", "chick", "click", "flick", "kick", "lick", "nick",
            "pick", "quick", "sick", "slick", "stick", "thick", "tick",
            "trick", "wick"],

    # -OCK sound (block, clock, rock)
    "ock": ["block", "clock", "cock", "dock", "flock", "knock", "lock", "mock",
            "rock", "shock", "sock", "stock", "unlock"],

    # -OT sound (dot, hot, spot)
    "ot": ["blot", "bot", "clot", "cot", "dot", "forgot", "got", "hot", "jot", "knot",
           "lot", "not", "plot", "pot", "robot", "rot", "shot", "slot", "snot", "spot",
           "squat", "trot", "yacht"],

    # -ELL sound (bell, cell, tell)
    "ell": ["bell", "cell", "dwell", "excel", "farewell", "fell", "hell", "hotel",
            "jell", "propel", "sell", "shell", "smell", "spell", "swell", "tell",
            "well", "yell"],

    # -ALL sound (ball, call, wall)
    "all": ["all", "ball", "call", "fall", "hall", "install", "mall", "recall", "small",
            "stall", "tall", "wall"],

    # -AND sound (band, hand, stand)
    "and": ["band", "brand", "command", "demand", "expand", "grand", "hand", "land",
            "sand", "stand", "understand"],

    # -ARK sound (dark, park, mark)
    # Note: removed "clerk" - pronounced "klerk" in American English
    "ark": ["ark", "bark", "dark", "embark", "landmark", "mark", "park",
            "remark", "shark", "spark", "stark"],

    # -ARM sound (arm, farm, harm)
    "arm": ["alarm", "arm", "charm", "disarm", "farm", "harm", "warm"],

    # -ART sound (art, cart, heart)
    "art": ["apart", "art", "cart", "chart", "dart", "depart", "heart", "impart",
            "mart", "part", "smart", "start", "tart"],

    # -AST sound (fast, last, past)
    "ast": ["blast", "broadcast", "cast", "contrast", "fast", "forecast", "last",
            "mast", "outcast", "past", "vast"],

    # -EAT sound (beat, eat, heat)
    "eat": ["beat", "cheat", "compete", "complete", "concrete", "defeat", "delete",
            "eat", "feat", "greet", "heat", "meat", "neat", "repeat", "seat", "sheet",
            "street", "sweet", "treat", "wheat"],

    # -OOL sound (cool, pool, school)
    "ool": ["cool", "cruel", "drool", "fool", "fuel", "jewel", "pool", "rule", "school",
            "spool", "stool", "tool"],

    # -ORE sound (core, more, store)
    "ore": ["adore", "before", "bore", "chore", "core", "door", "explore", "floor",
            "four", "ignore", "more", "pour", "restore", "roar", "score", "shore",
            "snore", "soar", "sore", "store", "swore", "tore", "war", "wore", "your"],

    # -OWN sound (brown, down, town)
    "own": ["brown", "clown", "crown", "down", "drown", "frown", "gown", "rown",
            "town", "unknown"],

    # -UMP sound (bump, jump, pump)
    "ump": ["bump", "clump", "dump", "grump", "hump", "jump", "lump", "plump", "pump",
            "slump", "stump", "trump", "thump"],

    # -UST sound (dust, just, must)
    "ust": ["adjust", "August", "bust", "crust", "disgust", "distrust", "dust",
            "entrust", "gust", "just", "must", "robust", "rust", "thrust", "trust"],

    # -INE sound (fine, line, mine)
    "ine": ["align", "assign", "combine", "confine", "decline", "define", "design",
            "dine", "divine", "fine", "incline", "line", "mine", "nine", "online",
            "outline", "pine", "refine", "resign", "shine", "shrine", "sign", "spine",
            "vine", "whine", "wine"],

    # -ICE sound (ice, nice, price)
    "ice": ["advice", "device", "dice", "ice", "mice", "nice", "price", "rice",
            "sacrifice", "slice", "spice", "splice", "suffice", "thrice", "twice",
            "vice"],

    # -IDE sound (guide, hide, ride)
    "ide": ["aside", "beside", "bride", "collide", "confide", "decide", "divide",
            "glide", "guide", "hide", "inside", "outside", "pride", "provide", "ride",
            "side", "slide", "stride", "tide", "wide"],

    # -IFE sound (knife, life, wife)
    "ife": ["knife", "life", "rife", "strife", "wife"],

    # -OSE sound (close, dose, those)
    "oze": ["close", "compose", "disclose", "dispose", "dose", "enclose", "expose",
            "froze", "glucose", "hose", "impose", "nose", "oppose", "pose", "propose",
            "prose", "rose", "suppose", "those"],

    # -OKE sound (broke, joke, smoke)
    "oke": ["awoke", "broke", "choke", "cloak", "coke", "evoke", "folk", "invoke",
            "joke", "oak", "poke", "provoke", "revoke", "smoke", "spoke", "stroke",
            "woke", "yoke"],

    # -OPE sound (hope, rope, scope)
    # Note: removed "envelope" (EN-vuh-lohp) - doesn't rhyme with hope
    "ope": ["cope", "elope", "grope", "hope", "mope", "pope", "rope",
            "scope", "slope"],

    # -OWL sound (bowl, control, soul)
    "ole": ["bowl", "coal", "console", "control", "enroll", "foal", "goal", "hole",
            "mole", "parole", "patrol", "pole", "poll", "role", "roll", "scroll",
            "sole", "soul", "stroll", "toll", "troll", "whole"],

    # -UE/OO sound (blue, clue, true) - words ending in clear "oo" sound
    # Note: removed "value", "venue", "virtue", "tissue", "rescue" - different ending sound
    "oo": ["accrue", "blue", "blew", "brew", "canoe", "chew", "clue", "construe",
           "crew", "cue", "debut", "dew", "drew", "due", "flew", "flu", "glue", "grew",
           "guru", "hue", "knew", "mildew", "nephew", "new", "pursue", "queue",
           "renew", "review", "screw", "shoe", "shrew", "slew", "stew",
           "sue", "tattoo", "threw", "through", "too", "true", "two", "view", "zoo"],

    # -OP sound (cop, drop, stop)
    # Note: removed "develop", "laptop", "workshop" - stress on different syllable
    "op": ["chop", "cop", "crop", "drop", "flop", "hop", "mop", "nonstop",
           "pop", "prop", "shop", "slop", "stop", "swap", "top"],

    # -AP sound (cap, gap, map)
    "ap": ["cap", "chap", "clap", "flap", "gap", "handicap", "lap", "map", "nap",
           "overlap", "rap", "scrap", "slap", "snap", "strap", "tap", "trap", "unwrap",
           "wrap", "zap"],

    # -IP sound (chip, clip, ship)
    "ip": ["chip", "clip", "dip", "drip", "equip", "flip", "grip", "hip", "lip",
           "membership", "ownership", "partnership", "relationship", "rip", "ship",
           "sip", "skip", "slip", "snip", "strip", "tip", "trip", "whip", "zip"],

    # -UP sound (cup, up)
    # Note: removed "corrupt", "disrupt", "erupt", "interrupt" - they end in "-upt"
    "up": ["backup", "breakup", "buildup", "checkup", "cleanup", "cup",
           "hiccup", "makeup", "markup", "pickup", "pup", "setup", "startup", "up"],

    # -AG sound (bag, flag, tag)
    "ag": ["bag", "brag", "drag", "flag", "gag", "lag", "nag", "rag", "sag", "shag",
           "slag", "snag", "stag", "swag", "tag", "wag", "zigzag"],

    # -OG sound (blog, dog, fog)
    "og": ["blog", "bog", "catalog", "clog", "cog", "dialog", "dog", "fog", "frog",
           "hog", "jog", "log", "smog"],

    # -UG sound (bug, drug, hug)
    "ug": ["bug", "debug", "drug", "dug", "hug", "jug", "mug", "plug", "rug", "shrug",
           "slug", "smug", "snug", "thug", "tug", "unplug"],

    # -EN sound (den, hen, ten)
    # Note: removed "again" (uh-GEN), kept words ending in clear -en sound
    "en": ["amen", "ben", "den", "eleven", "hen", "men", "oxygen", "pen", "seven",
           "ten", "then", "when", "zen"],

    # -UN sound (fun, gun, run)
    "un": ["begun", "bun", "done", "fun", "gun", "hun", "none", "nun", "one", "outdone",
           "outrun", "overrun", "pun", "run", "shun", "son", "spun", "stun", "sun",
           "ton", "undone", "won"],

    # -AN sound (ban, can, man)
    "an": ["ban", "began", "bran", "can", "clan", "fan", "japan", "man", "outran",
           "overran", "pan", "plan", "ran", "scan", "span", "tan", "than", "van"],

    # -IN sound (begin, in, win)
    # Note: removed "coin", "join", "ruin" - different sounds (-oin, -uin)
    "in": ["begin", "bin", "cabin", "captain", "chin", "din", "fin", "grin",
           "in", "inn", "kin", "margin", "muffin", "origin", "pin", "penguin",
           "pumpkin", "robin", "shin", "sin", "skin", "spin", "thin", "tin",
           "twin", "violin", "virgin", "vitamin", "win", "within"],

    # -ORN sound (born, corn, horn)
    "orn": ["acorn", "adorn", "born", "corn", "horn", "morn", "mourn", "popcorn",
            "scorn", "sworn", "thorn", "torn", "unicorn", "warn", "worn"],

    # -URN sound (burn, turn, learn)
    # Note: removed "pattern", "southern", "western", "saturn" - different vowel sound
    "urn": ["adjourn", "burn", "churn", "concern", "discern", "earn", "fern", "intern",
            "learn", "overturn", "return", "stern", "turn", "yearn"],

    # -OUND sound (around, found, ground)
    "ound": ["abound", "around", "astound", "background", "bound", "compound",
             "confound", "expound", "found", "ground", "hound", "mound", "playground",
             "pound", "profound", "rebound", "resound", "round", "sound", "surround",
             "underground", "wound"],

    # -OUD sound (cloud, crowd, loud)
    "oud": ["aloud", "cloud", "crowd", "loud", "proud", "shroud"],

    # -TION sound (action, nation, station)
    "shun": ["action", "addition", "admission", "ambition", "application", "attention",
             "auction", "aviation", "caution", "celebration", "collection", "combination",
             "communication", "competition", "concentration", "condition", "connection",
             "construction", "conversation", "cooperation", "corporation", "correction",
             "creation", "decision", "decoration", "definition", "demonstration",
             "description", "destination", "destruction", "detection", "determination",
             "direction", "discussion", "distribution", "division", "duration",
             "education", "election", "elevation", "emotion", "equation", "erosion",
             "eruption", "evolution", "exception", "execution", "exhibition",
             "expectation", "expedition", "explosion", "expression", "extension",
             "fashion", "fiction", "formation", "foundation", "fraction", "function",
             "generation", "graduation", "hesitation", "illustration", "imagination",
             "imitation", "immigration", "impression", "indication", "infection",
             "inflation", "information", "injection", "innovation", "inspection",
             "institution", "instruction", "intention", "interaction", "introduction",
             "invasion", "invention", "investigation", "invitation", "isolation",
             "junction", "legislation", "liberation", "limitation", "location",
             "mansion", "mention", "migration", "mission", "modification", "motion",
             "motivation", "nation", "navigation", "negotiation", "notion", "nutrition",
             "objection", "obligation", "observation", "occasion", "occupation",
             "ocean", "operation", "opinion", "opposition", "option", "organization",
             "orientation", "passion", "pension", "perception", "perfection",
             "permission", "petition", "pollution", "population", "portion", "position",
             "possession", "prediction", "preparation", "presentation", "preservation",
             "prevention", "prison", "probation", "production", "profession",
             "projection", "promotion", "proportion", "proposition", "protection",
             "provision", "publication", "question", "quotation", "radiation",
             "reaction", "reason", "reception", "recognition", "recommendation",
             "reduction", "reflection", "reformation", "refrigeration", "registration",
             "regulation", "rejection", "relation", "relaxation", "religion",
             "relocation", "renovation", "repetition", "representation", "reproduction",
             "reputation", "reservation", "resignation", "resolution", "restriction",
             "retention", "reunion", "revelation", "revolution", "rotation", "salvation",
             "sanction", "satisfaction", "season", "section", "selection", "sensation",
             "separation", "session", "situation", "solution", "specification",
             "speculation", "station", "stimulation", "subscription", "suggestion",
             "supervision", "suspension", "television", "temptation", "tension",
             "tradition", "transaction", "transformation", "transition", "translation",
             "transmission", "transportation", "tuition", "vacation", "variation",
             "vegetation", "version", "vibration", "violation", "vision", "vocation"],

    # -NESS sound (fitness, happiness, darkness)
    "ness": ["awareness", "blindness", "boldness", "brightness", "business", "careless",
             "calmness", "clearness", "cleverness", "closeness", "coldness", "coolness",
             "darkness", "deafness", "dizziness", "dryness", "dullness", "eagerness",
             "easiness", "effectiveness", "emptiness", "evenness", "exactness",
             "fairness", "fatness", "firmness", "fitness", "flatness", "fondness",
             "foolishness", "forgiveness", "frankness", "freshness", "friendliness",
             "fullness", "gentleness", "gladness", "goodness", "greatness", "greenness",
             "happiness", "hardness", "harshness", "highness", "holiness", "homeless",
             "hopeless", "illness", "kindness", "laziness", "likeness", "loneliness",
             "loudness", "madness", "meanness", "mildness", "misery", "nearness",
             "neatness", "nervousness", "newness", "oddness", "openness", "plainness",
             "politeness", "poorness", "powerless", "quickness", "quietness", "rawness",
             "readiness", "redness", "richness", "ripeness", "roughness", "roundness",
             "rudeness", "sadness", "sameness", "sharpness", "shyness", "sickness",
             "silence", "simpleness", "sleepiness", "slowness", "smoothness", "softness",
             "soreness", "soundness", "sourness", "stillness", "stiffness", "strangeness",
             "strictness", "stubbornness", "sweetness", "swiftness", "thickness",
             "thinness", "tightness", "tiredness", "ugliness", "uniqueness", "usefulness",
             "useless", "vagueness", "vastness", "warmness", "weakness", "wetness",
             "whiteness", "wholeness", "wideness", "wilderness", "wildness", "willingness",
             "witness", "worthless"],

    # Split -ER into actual rhyming subgroups based on stressed vowel before -er

    # -ETHER/-EATHER sound (together, weather, whether)
    "ether": ["feather", "leather", "together", "weather", "whether"],

    # -OTHER sound (mother, brother, other)
    "other": ["another", "brother", "mother", "other", "smother"],

    # -OVER sound (over, cover, discover)
    "over": ["cover", "discover", "hover", "over", "recover"],

    # -UNDER sound (under, thunder, wonder)
    "under": ["blunder", "plunder", "thunder", "under", "wonder"],

    # -EMBER sound (member, remember, december)
    "ember": ["december", "ember", "member", "november", "remember", "september"],

    # -ETTER sound (better, letter)
    "etter": ["better", "letter", "setter", "sweater"],

    # -UTTER sound (butter, shutter, utter)
    "utter": ["butter", "clutter", "flutter", "mutter", "shutter", "stutter", "utter"],

    # -IVER sound (river, deliver, silver)
    "iver": ["deliver", "liver", "quiver", "river", "shiver", "silver"],

    # -ORDER sound (order, border)
    "order": ["border", "disorder", "order", "recorder"],

    # -ASTER sound (master, disaster, faster)
    "aster": ["aster", "caster", "disaster", "faster", "master", "plaster"],

    # -AMMER sound (hammer, grammar)
    "ammer": ["glamour", "grammar", "hammer", "stammer"],

    # -ANGER sound (anger, danger, stranger)
    "anger": ["anger", "danger", "hanger", "ranger", "stranger"],

    # -INGER sound (finger, linger, singer)
    "inger": ["finger", "linger", "singer", "trigger"],

    # -UMMER sound (summer, drummer)
    "ummer": ["bummer", "drummer", "plumber", "slumber", "summer"],

    # -INNER sound (dinner, inner, winner)
    "inner": ["beginner", "dinner", "inner", "sinner", "spinner", "thinner", "winner"],

    # -OWER sound (flower, power, tower)
    "ower": ["devour", "flower", "hour", "power", "shower", "tower"],

    # -EVER sound (ever, never, forever)
    "ever": ["clever", "ever", "forever", "however", "lever", "never", "sever", "whatever",
             "whenever", "wherever", "whoever"],

    # -ARD sound (card, hard, guard)
    "ard": ["award", "backward", "billboard", "card", "discard", "forward", "guard",
            "hard", "hazard", "inward", "lizard", "outward", "regard", "reward",
            "standard", "steward", "toward", "upward", "wizard", "yard"],

    # -EEP sound (deep, keep, sleep)
    "eep": ["asleep", "beep", "cheap", "creep", "deep", "heap", "jeep", "keep", "leap",
            "peep", "reap", "seep", "sheep", "sleep", "steep", "sweep", "weep"],

    # -OOK sound (book, cook, look)
    "ook": ["book", "brook", "cook", "crook", "hook", "look", "mistook", "nook",
            "outlook", "overlook", "rook", "shook", "took", "undertook"],

    # -EW sound (few, new, view)
    "ew": ["blew", "brew", "chew", "crew", "dew", "drew", "few", "flew", "grew",
           "knew", "nephew", "new", "preview", "renew", "review", "screw", "shrew",
           "slew", "stew", "threw", "view"],

    # -OY sound (boy, enjoy, toy)
    "oy": ["alloy", "annoy", "boy", "convoy", "decoy", "deploy", "destroy", "employ",
           "enjoy", "joy", "ploy", "toy"],

    # -ANCE/-ENCE with same pronunciation
    "ance": ["advance", "alliance", "allowance", "ambulance", "appearance", "appliance",
             "assistance", "attendance", "balance", "brilliance", "chance", "circumstance",
             "clearance", "compliance", "conference", "confidence", "consequence",
             "convenience", "correspondence", "dance", "defense", "defiance", "difference",
             "distance", "dominance", "elegance", "endurance", "enhance", "entrance",
             "essence", "evidence", "excellence", "existence", "expense", "experience",
             "finance", "fragrance", "france", "glance", "governance", "guidance",
             "ignorance", "importance", "independence", "influence", "innocence",
             "instance", "insurance", "intelligence", "interference", "lance",
             "license", "magnificence", "maintenance", "negligence", "nuisance",
             "obedience", "observance", "occurrence", "offense", "patience", "performance",
             "permanence", "persistence", "preference", "presence", "prevalence",
             "prominence", "providence", "radiance", "reference", "relevance", "reliance",
             "reluctance", "remembrance", "renaissance", "repentance", "resemblance",
             "residence", "resistance", "resonance", "romance", "sentence", "sequence",
             "significance", "silence", "stance", "substance", "surveillance",
             "tolerance", "trance", "utterance", "variance", "violence"],

    # -ENCE with different pronunciation (like fence, hence)
    "ense": ["commence", "condense", "defense", "dense", "dispense", "expense", "fence",
             "hence", "immense", "incense", "intense", "nonsense", "offense", "pretense",
             "sense", "suspense", "tense", "thence", "whence"],
}

def main():
    # Read all BIP39 words
    with open(os.path.join(os.path.dirname(__file__), '..', 'data', 'english.txt')) as f:
        all_words = set(line.strip().lower() for line in f if line.strip())

    # Filter rhyme groups to only include BIP39 words
    filtered_groups = {}
    for group_name, words in PHONETIC_RHYME_GROUPS.items():
        bip39_words = [w for w in words if w in all_words]
        if len(bip39_words) >= 2:  # Need at least 2 words to rhyme
            filtered_groups[group_name] = bip39_words

    data = datafile.load()
    data["rhymes"] = [[name, sorted(words)] for name, words in sorted(filtered_groups.items())]
    datafile.save(data)

    # Stats
    total_words = sum(len(words) for words in filtered_groups.values())
    print(f"Updated data/seedpoet.json:")
    print(f"  - {len(filtered_groups)} rhyme groups")
    print(f"  - {total_words} total word entries")

    # Show groups
    for name, words in sorted(filtered_groups.items(), key=lambda x: -len(x[1])):
        print(f"  - {name}: {len(words)} words")

if __name__ == '__main__':
    main()
