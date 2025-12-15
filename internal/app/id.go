package app

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

var adjectives = []string{
	"adorable", "adventurous", "aggressive", "agreeable", "alert", "alive", "amused", "angry",
	"annoyed", "annoying", "anxious", "arrogant", "ashamed", "attractive", "average", "awful",
	"bad", "beautiful", "better", "bewildered", "black", "bloody", "blue", "blue-eyed", "blushing",
	"bored", "brainy", "brave", "breakable", "bright", "busy", "calm", "careful", "cautious",
	"charming", "cheerful", "clean", "clear", "clever", "cloudy", "clumsy", "colorful", "combative",
	"comfortable", "concerned", "condemned", "confused", "cooperative", "courageous", "crazy", "creepy",
	"crowded", "cruel", "curious", "cute", "dangerous", "dark", "dead", "defeated", "defiant",
	"delightful", "depressed", "determined", "different", "difficult", "disgusted", "distinct", "disturbed",
	"dizzy", "doubtful", "drab", "dull", "eager", "easy", "elated", "elegant", "embarrassed",
	"enchanting", "encouraging", "energetic", "enthusiastic", "envious", "evil", "excited", "expensive",
	"exuberant", "fair", "faithful", "famous", "fancy", "fantastic", "fierce", "filthy",
	"fine", "foolish", "fragile", "frail", "frantic", "friendly", "frightened", "funny",
	"gentle", "gifted", "glamorous", "gleaming", "glorious", "good", "gorgeous", "graceful",
	"grieving", "grotesque", "grumpy", "handsome", "happy", "healthy", "helpful", "helpless",
	"hilarious", "homeless", "homely", "horrible", "hungry", "hurt", "ill", "important",
	"impossible", "inexpensive", "innocent", "inquisitive", "itchy", "jealous", "jittery", "jolly",
	"joyous", "kind", "lazy", "light", "lively", "lonely", "long", "lovely",
	"lucky", "majestic", "magnificent", "misty", "modern", "motionless", "muddy", "mushy",
	"mysterious", "nasty", "naughty", "nervous", "nice", "nutty", "obedient", "obnoxious",
	"odd", "old-fashioned", "open", "outrageous", "outstanding", "panicky", "perfect", "plain",
	"pleasant", "poised", "poor", "powerful", "precious", "prickly", "proud", "puzzled",
	"quaint", "real", "relieved", "repulsive", "rich", "scary", "selfish", "shiny",
	"shy", "silly", "sleepy", "smiling", "smoggy", "sore", "sparkling", "splendid",
	"spotless", "stormy", "strange", "stupid", "successful", "super", "talented", "tame",
	"tasty", "tender", "tense", "terrible", "thankful", "thoughtful", "thoughtless", "tired",
	"tough", "troubled", "ugliest", "ugly", "uninterested", "unsightly", "unusual", "upset",
	"uptight", "vast", "victorious", "vivacious", "wandering", "weary", "wicked", "wide-eyed",
	"wild", "witty", "worried", "worrisome", "wrong", "zany", "zealous",
}

var animals = []string{
	"aardvark", "albatross", "alligator", "alpaca", "ant", "anteater", "antelope", "ape",
	"armadillo", "baboon", "badger", "barracuda", "bat", "bear", "beaver", "bee",
	"bison", "boar", "buffalo", "butterfly", "camel", "capybara", "caribou", "cassowary",
	"cat", "caterpillar", "cattle", "chamois", "cheetah", "chicken", "chimpanzee", "chinchilla",
	"chough", "clam", "cobra", "cockroach", "cod", "cormorant", "coyote", "crab",
	"crane", "crocodile", "crow", "curlew", "deer", "dinosaur", "dog", "dogfish",
	"dolphin", "donkey", "dotterel", "dove", "dragonfly", "duck", "dugong", "dunlin",
	"eagle", "echidna", "eel", "eland", "elephant", "elk", "emu", "falcon",
	"ferret", "finch", "fish", "flamingo", "fly", "fox", "frog", "gaur",
	"gazelle", "gerbil", "giraffe", "gnat", "gnu", "goat", "goldfinch", "goldfish",
	"goose", "gorilla", "goshawk", "grasshopper", "grouse", "guanaco", "gull", "hamster",
	"hare", "hawk", "hedgehog", "heron", "herring", "hippopotamus", "hornet", "horse",
	"hummingbird", "hyena", "ibex", "ibis", "jackal", "jaguar", "jay", "jellyfish",
	"kangaroo", "kingfisher", "koala", "komodo", "kookabura", "kouprey", "kudu", "lapwing",
	"lark", "lemur", "leopard", "lion", "llama", "lobster", "locust", "loris",
	"louse", "lyrebird", "magpie", "mallard", "manatee", "mandrill", "mantis", "marten",
	"meerkat", "mink", "mole", "mongoose", "monkey", "moose", "mosquito", "mouse",
	"mule", "narwhal", "newt", "nightingale", "octopus", "okapi", "opossum", "oryx",
	"ostrich", "otter", "owl", "oyster", "panther", "parrot", "partridge", "peafowl",
	"pelican", "penguin", "pheasant", "pig", "pigeon", "pony", "porcupine", "porpoise",
	"quail", "quelea", "quetzal", "rabbit", "raccoon", "rail", "ram", "rat",
	"raven", "reindeer", "rhino", "rook", "salamander", "salmon", "sandpiper", "sardine",
	"scorpion", "seahorse", "seal", "shark", "sheep", "shrew", "skunk", "snail",
	"snake", "sparrow", "spider", "spoonbill", "squid", "squirrel", "starling", "stingray",
	"stinkbug", "stork", "swallow", "swan", "tapir", "tarsier", "termite", "tiger",
	"toad", "trout", "turkey", "turtle", "viper", "vulture", "wallaby", "walrus",
	"wasp", "weasel", "whale", "wildcat", "wolf", "wolverine", "wombat", "woodcock",
	"woodpecker", "worm", "wren", "yak", "zebra",
}

func GetIdentity(peerId peer.ID) string {
	idString := peerId.String()

	hash := sha256.Sum256([]byte(idString))

	adjectiveIndex := binary.BigEndian.Uint64(hash[0:8]) % uint64(len(adjectives))
	animalIndex := binary.BigEndian.Uint64(hash[8:16]) % uint64(len(animals))
	number := binary.BigEndian.Uint64(hash[16:24]) % 1000

	return fmt.Sprintf("@%s-%s-%d", adjectives[adjectiveIndex], animals[animalIndex], number)
}
