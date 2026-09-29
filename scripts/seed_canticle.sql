-- ==========================================================================
-- Canticle of the End — Campaign Seed Data
-- Campaign ID: 36 (Canticle Of The End, CoC 7e)
--
-- Run with: psql -f scripts/seed_canticle.sql
-- Or via docker: docker exec -i imagineer-postgres psql -U imagineer -f -
-- ==========================================================================

BEGIN;

-- Verify campaign exists
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM campaigns WHERE id = 36) THEN
    RAISE EXCEPTION 'Campaign 36 does not exist. Create it first.';
  END IF;
END $$;

-- ==========================================================================
-- CHAPTERS
-- ==========================================================================

INSERT INTO chapters (campaign_id, title, overview, sort_order) VALUES
(36, 'London — The Orphean Society',
 'Chapter 1 introduces the investigators to the Aeternum Choir through the Orphean Society, a cult hiding behind public respectability in London''s musical salons. Investigation leads from London society to the Stonehenge ritual site, where the Choir intends to complete the core harmonic framework. Dates: Early June 1814.',
 0),
(36, 'Lyon — The Societe Harmonique de l''Aube',
 'Chapter 2 escalates the horror from London''s genteel salons into something far more visceral. The Societe Harmonique de l''Aube operates at the intersection of post-Revolutionary French esotericism and pre-Christian Gallo-Roman harmonic tradition, using the acoustic properties of Roman ruins beneath Fourviere as a resonance chamber. Ritual date: July 14 (Bastille Day). Dates: Late June to Mid-July 1814.',
 1),
(36, 'Venice — The Confraternita del Bel Canto',
 'An operatic cult posing as a patronage guild, using sound geometry in canal acoustics and emotional devastation through aria. This chapter is resolved off-camera by Varrio Harrowmont, who disrupts the cell from within before joining the main party in Vienna. Dates: Late July to Early August 1814.',
 2),
(36, 'Vienna — The Brotherhood of the Open Measure',
 'Vienna is the campaign''s structural pivot. The Brotherhood believes humanity''s "error" must be eliminated to achieve divine harmony. Their horror is the Harmonic Engine — a 30-foot biomechanical instrument built from human body parts. The Congress of Vienna provides the social backdrop. Dates: August 3-15, 1814. Ritual: Midnight, August 15 (Feast of the Assumption).',
 3),
(36, 'Calcutta',
 'The final player chapter. Calcutta in 1814 is the capital of British India. The cult operates at the intersection of British colonial society and Indian mystical tradition. The ritual is built around funerary chants and river resonance, exploiting the Hooghly River''s acoustic properties. Ritual date: Late October 1814 (Kali Puja / monsoon break).',
 4);

-- ==========================================================================
-- ENTITIES — Player Characters
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Marina Garrick', 'pc',
 'Surrey gentlewoman, survivor of Osney Grange. Perhaps the most committed to the Order''s mission.',
 'Aspiring investigator, growing in confidence. Haunted by past horrors.',
 ARRAY['investigator', 'order-of-st-aelfric', 'british', 'gentlewoman'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Emma Wentworth', 'pc',
 'Wiltshire gentlewoman, theatre-obsessed since the Drury Lane incident. Social charm, performance skills, emotional heart of the group.',
 'Courageous, dramatic, loyal to her sister Georgiana.',
 ARRAY['investigator', 'order-of-st-aelfric', 'british', 'gentlewoman', 'performer'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Georgiana Wentworth', 'pc',
 'Emma''s sister, sharp-minded analyst. Intelligence, deduction, and research specialist.',
 'Phobias of theatre and fish (developed from "The Long Corridor"). Her theatre phobia creates tension in a city obsessed with performance.',
 ARRAY['investigator', 'order-of-st-aelfric', 'british', 'gentlewoman', 'analyst'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Charlotte Thorne', 'pc',
 'Northumberland heiress, steady and skeptical. Voice of reason with financial resources and social standing.',
 'Perceptive, skeptical of the uncanny, grounded.',
 ARRAY['investigator', 'order-of-st-aelfric', 'british', 'heiress'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Colonel Henri Moreau', 'pc',
 'French military officer, Napoleonic Wars veteran. Combat expertise and Continental perspective.',
 'A French officer in Vienna in 1814 is politically complex — France just lost the war. Austrian officials may view him with suspicion. Replaced Augustus Bolt.',
 ARRAY['investigator', 'order-of-st-aelfric', 'french', 'military', 'officer'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Varrio Harrowmont', 'pc',
 'Italian gentleman, Order of St. Aelfric member. Took down the Venice cult (Confraternita del Bel Canto) before joining the main party in Lyon.',
 'Experienced cult-hunter. Knows what the Aeternum Choir is capable of. Has seen the effects of partial Canticle exposure — singers whose voices no longer sound human. Italian nationals viewed with suspicion in Austrian Vienna.',
 ARRAY['investigator', 'order-of-st-aelfric', 'italian', 'cult-hunter'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md');

-- ==========================================================================
-- ENTITIES — Order of St. Aelfric / Allied NPCs
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Lord Percival Harcourt', 'npc',
 'Earl of Wrexham, the Order of St. Aelfric''s senior operative. The investigators'' employer who coordinates the broader campaign against the Aeternum Choir.',
 'In Vienna as official British diplomatic observer for the Congress. Does NOT know the investigators are in the city until the Imperial Reception reunion on August 6. After learning the scope of the Choir''s network, dispatches Order agents to remaining active cells.',
 ARRAY['order-of-st-aelfric', 'british', 'aristocrat', 'diplomat', 'earl', 'ally'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Lady Honoria Lyndhurst', 'npc',
 'Harcourt''s companion and the investigators'' original trainer. Present in Vienna for the Congress season.',
 'Neither she nor the investigators know each other is in Vienna — creates dramatic reunion opportunity. Provides continuity and emotional anchor for the party.',
 ARRAY['order-of-st-aelfric', 'british', 'aristocrat', 'ally', 'trainer'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Dr. Ernst Falkner', 'npc',
 'Austrian geologist and natural philosopher at the Naturhistorisches Institut, age 52. Not a formal Order member but a sympathetic contact.',
 'Lost his daughter Margarethe to a "musical academy" 3 years ago — she vanished and may be part of the Harmonic Engine. Knows Herzfeld is dangerous, that the University anatomical theatre was unsealed, and about the pattern of missing musicians. Code phrase: "the librarian of Saint-Just". Can provide lodging advice, University access/scholar credentials, and warning about August 15 timing.',
 ARRAY['ally', 'austrian', 'scientist', 'geologist', 'contact', 'order-sympathizer'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Dr. Wilhelm Brenner', 'npc',
 'Viennese surgeon, age 41. Former Brotherhood member who participated in early Engine experiments. Now hiding in Leopoldstadt, drinking himself to death while trying to find courage to expose Herzfeld.',
 'STR 45 CON 40 SIZ 60 DEX 55 APP 35 INT 75 POW 35 EDU 80 SAN 28 HP 10. Medicine 75%, Surgery 70%. Phobia: cannot bear sustained musical tones. Addiction: brandy. Delusion: hears "the voice of the Engine." Hiding at The Crooked Chimney tavern. Cult hunting him — Adler has hired criminals to watch Leopoldstadt. Days 1-3: safe. Days 4-6: located. Days 7-9: attempts to contact investigators. Days 10-12: captured or killed unless intervened.',
 ARRAY['defector', 'informant', 'austrian', 'surgeon', 'former-brotherhood', 'doomed'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Thomas Wyndham', 'npc',
 'Keeper-controlled NPC, loyal but volatile. Travels with the investigators.',
 'Jealousy thread with Graf von Sternberg creates romantic complication. May be provoked into a duel.',
 ARRAY['british', 'companion', 'volatile', 'keeper-npc'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Madame Celestine Delacroix', 'npc',
 'French widow, age 32, whose husband died at the Battle of Leipzig in 1813. Travelling to Vienna to petition for a widow''s pension.',
 'APP 70 INT 65 Charm 55% Persuade 50%. Encountered at the Linienwall customs gate — damsel in distress. If helped, provides her Vienna address and mentions she has a letter of introduction to Countess von Thun (her late husband''s godmother). This creates an organic path into high society.',
 ARRAY['french', 'widow', 'ally', 'social-connection', 'customs-scene'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md');

-- ==========================================================================
-- ENTITIES — Brotherhood of the Open Measure (Vienna Cult)
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Professor Albin Herzfeld', 'npc',
 'Music theorist, anatomist, and former Imperial Court mathematician, age 47. Professor at the University of Vienna and leader of the Brotherhood of the Open Measure. Designer of the Harmonic Engine.',
 'Cold, clinical, genuinely believes he is serving a higher purpose. Tall, gaunt, wire-rimmed spectacles, ink-stained fingers. Non-combatant — will flee or surrender. Final words if captured: "You''ve destroyed the first perfect instrument... but others will build again." After Session 4 events: fully aware hostile agents are operating in Vienna. Has ratcheted security to maximum.',
 ARRAY['cult-leader', 'austrian', 'professor', 'antagonist', 'brotherhood', 'scientist'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Kapellmeister Friedrich Adler', 'npc',
 'Conductor and composer, age 35. Cult lieutenant who recruits "volunteers" and identifies suitable "components" for the Harmonic Engine.',
 'Fanatically loyal to Herzfeld. Skilled with knives, will fight to death. Secret: his sister''s voice is part of the Engine. Tasked with silencing Brenner permanently — has hired local criminals to watch Leopoldstadt.',
 ARRAY['cult-lieutenant', 'austrian', 'musician', 'antagonist', 'brotherhood', 'dangerous'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Count Leopold von Trauttmansdorff', 'npc',
 'Wealthy noble, age 56. Cult financier providing funding and political protection to the Brotherhood.',
 'Believes the ritual will cure his chronic illness. A coward who will betray others if caught. Knows names, locations, and financial records. Key interrogation target if captured.',
 ARRAY['cult-financier', 'austrian', 'aristocrat', 'antagonist', 'brotherhood', 'weak-link'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Baron Otto von Kaunitz', 'npc',
 'Aristocrat, age 38. Secret cult sympathizer (not full member) who reports investigators'' movements to Herzfeld.',
 'Cover: patron of musical innovation. Smooth, charming, dead eyes. Has eyes everywhere — any approach to the University district is observed and reported. The party cannot scout without being noticed.',
 ARRAY['cult-sympathizer', 'austrian', 'aristocrat', 'antagonist', 'brotherhood', 'spy'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Inspektor Heinrich Vogel', 'npc',
 'Geheimpolizei officer and secret Brotherhood member. A true believer in the cult''s mission.',
 'Deploys uniformed officers to the University perimeter under cover of "protecting state property during the Congress." Any assault on the University becomes an assault on Austrian police.',
 ARRAY['cult-member', 'austrian', 'police', 'antagonist', 'brotherhood', 'geheimpolizei'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Anna Lindqvist', 'npc',
 'Soprano selected as the final vocal component of the Harmonic Engine. Her voice is intended to complete Segment V.',
 'The cult may have a backup soprano if the PCs rescue her. One of the key moral dilemmas: save Anna or stop the ritual?',
 ARRAY['victim', 'soprano', 'engine-component', 'brotherhood'],
 'AUTHORITATIVE', 'Campaign_Design_Notes.md');

-- ==========================================================================
-- ENTITIES — Viennese Society & Diplomats
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Countess Maria von Thun', 'npc',
 'Vienna''s most influential society hostess, age 52. Elegant, sophisticated patron of the arts who can make or break reputations.',
 'Sharp-witted, knows everyone''s secrets. Neutral — horrified by cult if she learns truth. Can provide introductions and social intelligence. The investigators need her salon invitation to access high society.',
 ARRAY['austrian', 'aristocrat', 'society', 'patron', 'hostess'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Princess Esterhazy', 'npc',
 'Renowned beauty, age 34. Leader of Vienna''s fashionable set who controls social access.',
 'Noticed "odd musicians" at private concerts. May pursue male investigators romantically.',
 ARRAY['austrian', 'aristocrat', 'society', 'fashionable'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Prince Klemens von Metternich', 'npc',
 'Austrian Foreign Minister, age 41. The most powerful man in Vienna, organizing the Congress. Runs the Geheimpolizei secret police.',
 'Brilliant, calculating, elegant, suspicious of British motives. If mishandled, could have investigators arrested as spies. Potential ally only if presented with ironclad evidence carefully.',
 ARRAY['austrian', 'politician', 'diplomat', 'foreign-minister', 'powerful', 'historical'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Graf Heinrich von Reichenbach', 'npc',
 'Minor Austrian nobility, age 35. Imperial court official and secret composer writing music under the pseudonym "Heinrich Schiller."',
 'Widower (wife died in childbirth). Melancholic, artistic, trapped by duty. Knows about Herzfeld''s "innovations" and is disturbed. Can provide insider access to University and court circles. Romantic interest for female investigators.',
 ARRAY['austrian', 'aristocrat', 'court-official', 'composer', 'romantic-interest'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Graf Maximilian von Sternberg', 'npc',
 'Austrian cavalry officer, Fourth Hussars. Handsome, arrogant, impeccably connected. Potential romantic rival for Captain Wyndham.',
 'Could insult Wyndham publicly, forcing a duel. May have ties to the Brotherhood (manipulation or coincidence?). Duel subplot drives dramatic tension.',
 ARRAY['austrian', 'military', 'officer', 'romantic-rival', 'hussars'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Count Karl von Nesselrode', 'npc',
 'Russian diplomat representing Tsar Alexander I at the Congress.',
 'Skeptical of Austrian occultism, might believe cult story. Knows about disappearances of Russian musicians in Vienna.',
 ARRAY['russian', 'diplomat', 'congress'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Baron Wilhelm von Humboldt', 'npc',
 'Prussian scholar-diplomat and representative at the Congress. Interested in natural philosophy.',
 'Would be horrified by Herzfeld''s perversion of science. Has noticed statistical anomalies in missing persons. Bound by diplomatic protocol.',
 ARRAY['prussian', 'diplomat', 'scholar', 'congress', 'historical'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Count Nikolai Volkonsky', 'npc',
 'Russian cavalry officer, age 28, attached to the Russian delegation. Tall, fair-haired, ice-blue eyes, dueling scar on left cheek.',
 'Romantic, passionate, slightly melancholic. Approach: poetry, intense conversations about fate. Expected to marry Russian nobility. Drinks heavily, haunted by war, impulsive. Has noticed disappearances among Russian musicians.',
 ARRAY['russian', 'military', 'officer', 'romantic-interest', 'congress'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Signore Lorenzo di Firenze', 'npc',
 'Minor Italian nobility, age 26. Poet and dilettante in Vienna for the Congress season.',
 'Passionate, artistic, dramatic, genuine beneath the performance. Approach: sonnets, serenades, grand romantic declarations. Has connections in Vienna''s artistic underground.',
 ARRAY['italian', 'poet', 'aristocrat', 'romantic-interest'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Countess Katerina Orlova', 'npc',
 'Minor Russian nobility, age 24. Lady-in-waiting to the Russian delegation.',
 'Romantic interest for male investigators. Further details to be developed.',
 ARRAY['russian', 'aristocrat', 'romantic-interest', 'congress'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Dorothea de Courlande', 'npc',
 'Talleyrand''s niece, French intelligence conduit at the Congress of Vienna.',
 NULL,
 ARRAY['french', 'diplomat', 'intelligence', 'congress'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Count de la Tour-du-Pin', 'npc',
 'French delegation member at the Congress. Provides intelligence on missing musicians and cult movements.',
 NULL,
 ARRAY['french', 'diplomat', 'intelligence', 'congress'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md');

-- ==========================================================================
-- ENTITIES — University of Vienna NPCs
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Herr Benedikt Kuscher', 'npc',
 'Head Librarian of the University of Vienna, age 60. Precise, territorial man with the bearing of a former Benedictine monk.',
 'Knows every volume in his domain, suspicious of irregular requests. Will provide access to visiting scholars with proper credentials. Has noticed increased interest in anatomical works lately.',
 ARRAY['austrian', 'librarian', 'university', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Professor Leopold Karwinsky', 'npc',
 'Rector of the University of Vienna, age 58. Natural philosopher more interested in administration than imagination.',
 'Cooperative with authority, suspicious of unsanctioned inquiries.',
 ARRAY['austrian', 'rector', 'university', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Wolfgang Staufer', 'npc',
 'Master Porter of the medical faculty at the University of Vienna, age 68. Forty-three years in his position — knows every secret passage.',
 'Dark coat worn smooth with age, enormous ring of keys, gambler''s face, faded blue eyes. Wary, shrewd, can be bribed (drinks and needs money). Knows everything about the sealed theatre including Herzfeld''s activities. Recently nervous, jumpy, drinking more heavily. Has keys to the anatomical theatre.',
 ARRAY['austrian', 'porter', 'university', 'bribable', 'key-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Herr Karl Fenster', 'npc',
 'Chief administrative clerk at the University, age 52. Maintains elaborate ledgers of all university property and facilities.',
 'Officious, meticulous. Knows the theatre was sealed in 1794 for "unfortunate incidents." Does NOT know about Herzfeld''s reopening — will be disturbed to learn of it.',
 ARRAY['austrian', 'clerk', 'university', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Johann Moeller', 'npc',
 'Thin, nervous third-year medical student at the University, age 24. Witnessed Herzfeld and another man in the sealed theatre two weeks ago.',
 'Terrified. Convinced he witnessed necromancy. Afraid of Herzfeld''s influence over his academic future. Genuine sympathy gets him to talk; threats make him clam up.',
 ARRAY['austrian', 'student', 'university', 'witness', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md');

-- ==========================================================================
-- ENTITIES — Victims & Survivors
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Frau Margarethe Holzer', 'npc',
 'Former opera singer, age 33. Willing donor who gave her left lung and voice box to the Harmonic Engine and survived.',
 'Speaks in whispers, grotesquely scarred, fervent believer. Quote: "I am eternal now... my voice sings in the divine mechanism..." Encountered at cult gatherings.',
 ARRAY['victim', 'donor', 'opera-singer', 'believer', 'brotherhood'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Hans Gruber', 'npc',
 'Failed pianist, age 28. Donated both hands to the Engine and survived. Arms end at wrists with hook prosthetics.',
 'Bitter and broken. Could be convinced to testify against cult. Knows Engine''s location and weaknesses.',
 ARRAY['victim', 'donor', 'pianist', 'bitter', 'informant'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Margarethe Falkner', 'npc',
 'Dr. Falkner''s daughter, selected as a "soprano component" for the Harmonic Engine. Vanished 3 years ago after joining a "musical academy."',
 'Brenner doesn''t know if she is still alive. If she is, she is part of the Engine. He remembers her audition — she had a beautiful voice.',
 ARRAY['victim', 'engine-component', 'soprano', 'missing'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md');

-- ==========================================================================
-- ENTITIES — Lyon Chapter NPCs
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Mathilde Savarin', 'npc',
 'Composer, patron of the arts, and high priestess of the Societe Harmonique de l''Aube. Strikingly beautiful in a haunting, statuesque way, age 46.',
 'Height 5''10", pale violet-gray eyes, glossy black hair in Greco-Roman coils with pearl pins. Wears white, grey, and muted violet with musical notation or thorn motif embroidery. Voice: low, precise, melodic. Presence: impossible to ignore. Cult leader of the Lyon cell.',
 ARRAY['cult-leader', 'french', 'composer', 'antagonist', 'societe-harmonique'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Captain Luc Fouchard', 'npc',
 'Former artillery captain in Napoleon''s army, age 39. Now commander of Savarin''s Blue Sash enforcers — the Societe Harmonique''s militant wing.',
 'Broad-shouldered, square-jawed, weathered. Short military crop flecked with gray. Carries a sabre. Wears the blue sash with three interlocked rings sigil. Controlled violence under military discipline. Believes utterly in Savarin''s philosophy. Has personally killed dissenters. Replaced Captain Henri Valois (deleted from continuity).',
 ARRAY['cult-lieutenant', 'french', 'military', 'antagonist', 'societe-harmonique', 'enforcer'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Duchesse Camille de Brissac', 'npc',
 'Lyon society hostess who hosts the July 3rd ball at Hotel de Brissac.',
 'Royalist. Razor politeness. Demands composure during crises.',
 ARRAY['french', 'aristocrat', 'society', 'lyon'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Abbe Duplessis', 'npc',
 'Cleric present at Lyon social events. Provides discreet warnings about Savarin''s activities.',
 'Hints about excavations beneath the city and tremors.',
 ARRAY['french', 'cleric', 'ally', 'lyon'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Jean-Luc Duret', 'npc',
 'Prefect of Lyon. Government official managing civil order during turbulent post-Napoleonic period.',
 'Tolerates Fouchard but dislikes his private patrols.',
 ARRAY['french', 'politician', 'prefect', 'lyon'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Jules Delaroche', 'npc',
 'Commissaire Principal of the Lyon Gendarmerie. Head of police in Lyon.',
 'Politically trapped between Prefect Duret and Savarin''s influence. Loyal to civic duty but disturbed by Savarin''s growing private power. Potential ally — may discreetly pass information. Originally duplicated as "Commissaire Renard" which was deleted.',
 ARRAY['french', 'police', 'ally', 'lyon', 'gendarmerie'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Elise Fontaine', 'npc',
 'Violinist who performs at Lyon social events. Notices strange changes in musical pitch and acoustics.',
 'Whispers about "the voices beneath the strings." A bowstring snaps during performance — she gasps "It happens again."',
 ARRAY['french', 'musician', 'violinist', 'lyon', 'minor-npc'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Doctor Carreau', 'npc',
 'Doctor in Lyon involved with Savarin''s acoustic experiments. His instruments cause resonance trauma.',
 'A cracked glass ampoule of his design causes a servant''s collapse at the ball — inner-ear trauma from intense vibration.',
 ARRAY['french', 'doctor', 'antagonist', 'lyon', 'societe-harmonique'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt');

-- ==========================================================================
-- ENTITIES — Minor Vienna NPCs
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Frau Dorothea Hofmann', 'npc',
 'Elderly widow who runs Pension Hofmann, an academic boarding house near the University in the Alsergrund district.',
 'Gossips. 1-2 gulden per night. Near University — good for investigation but less respectable address.',
 ARRAY['austrian', 'landlady', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Fraulein Elise von Schonberg', 'npc',
 'Countess von Thun''s widowed niece, staying in Vienna to recover from her husband''s death. Frequents quiet coffee houses.',
 'Prefers intelligent conversation to society''s circus. A genuine friendship or romantic connection with an investigator naturally leads to an introduction to her aunt. Takes several days to build.',
 ARRAY['austrian', 'aristocrat', 'widow', 'romantic-interest', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Inspector Novak', 'npc',
 'Austrian customs official at the Linienwall gates. Thorough but not hostile.',
 'Complications at customs: may find Herzfeld''s letter, undeclared weapons, or suspicious occult texts.',
 ARRAY['austrian', 'customs', 'minor-npc'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md');

-- ==========================================================================
-- ENTITIES — Deity
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Yog-Sothoth', 'deity',
 'The Harmonic Totality — the being whose vibrations compose reality. Known to the Aeternum Choir as The Chord Beyond the Veil, The Grand Interval, That Which Resonates All Things, and The Cantor of All Time.',
 'Not a god in the traditional sense but the living principle of harmonic unity. All of creation is sound in suspension. Mortals exist in a flawed octave, deaf to higher harmonics. The Grand Canticle intends to open all minds to Yog-Sothoth, erasing dissonance at the cost of individuality, time, and form. The result is oblivion disguised as revelation.',
 ARRAY['outer-god', 'cosmic-horror', 'cthulhu-mythos', 'harmonic'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md');

-- ==========================================================================
-- ENTITIES — Organizations
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Aeternum Choir', 'cult',
 'A globe-spanning cult that seeks cosmic transcendence through perfect harmony. Their goal is transfiguration — to raise humanity beyond mundane perception by invoking Yog-Sothoth through the Grand Canticle performed simultaneously at eight sacred sites.',
 'Requires 5 of 8 ritual segments to succeed before summer solstice 1815. Three cells already destroyed (London, Lyon, Venice). Five remain active (Vienna, Warsaw, Luxor, Calcutta, Chengdu/Ouro Preto). Coordinated through coded correspondence by "Der Kantor" based in Munich.',
 ARRAY['cult', 'global', 'musical', 'yog-sothoth'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Order of St. Aelfric', 'organization',
 'Anti-Mythos organization that employs the investigators. Led by Lord Percival Harcourt, Earl of Wrexham.',
 'After the Vienna reunion, transforms from ad hoc investigation to coordinated global counter-operation. Dispatches agents to remaining active cells.',
 ARRAY['anti-mythos', 'british', 'secret-society', 'protagonists'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Brotherhood of the Open Measure', 'cult',
 'The Vienna cell of the Aeternum Choir. Led by Professor Herzfeld. Believes humanity''s "error" must be eliminated to achieve divine harmony through mechanizing performance entirely.',
 'Approximately 15-20 full members plus servants and unwitting accomplices. Responsible for the Harmonic Engine beneath the University of Vienna. Ritual: Segment V — Harmonic Engine, midnight August 15.',
 ARRAY['cult', 'vienna', 'brotherhood', 'aeternum-choir-cell'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Societe Harmonique de l''Aube', 'cult',
 'The Lyon cell of the Aeternum Choir. Led by Mathilde Savarin. Operates at the intersection of post-Revolutionary French esotericism and pre-Christian Gallo-Roman harmonic tradition.',
 'Repurposes Gregorian chant and proto-Cathar mysticism into blasphemous ritual. Uses acoustic properties of Roman ruins beneath Fourviere. Ritual: Segment II — Liturgical Bridge Movement, July 14. Has a militant wing: the Blue Sash enforcers led by Captain Fouchard.',
 ARRAY['cult', 'lyon', 'french', 'aeternum-choir-cell'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Confraternita del Bel Canto', 'cult',
 'The Venice cell of the Aeternum Choir. An operatic cult posing as a patronage guild, using sound geometry in canal acoustics.',
 'Disrupted by Varrio Harrowmont before the main party reached Vienna. Segment III — The Soprano Ascension. Herzfeld''s correspondence notes "the Venice cell was stopped from within by a traitor."',
 ARRAY['cult', 'venice', 'italian', 'aeternum-choir-cell', 'destroyed'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Orphean Society', 'cult',
 'The London cell of the Aeternum Choir. A cult hiding behind public respectability in London''s musical salons.',
 'Destroyed by the party in June 1814. Segment I — Core Harmonic Framework. Ritual site: Stonehenge.',
 ARRAY['cult', 'london', 'british', 'aeternum-choir-cell', 'destroyed'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Geheimpolizei', 'organization',
 'Metternich''s secret police. Pervasive surveillance apparatus in Vienna that opens and copies mail, monitors foreigners, has informants in every major household.',
 'Inspektor Vogel is a Brotherhood mole within the Geheimpolizei. Creates paranoia — are watchers cultists, police, or both?',
 ARRAY['austrian', 'police', 'secret-police', 'surveillance'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Congress of Vienna', 'event',
 'International diplomatic gathering in Vienna, officially opening September 1814. Diplomats, nobles, and courtiers arriving throughout August. Provides the social backdrop and cover for cult activities.',
 '"The Congress will be dancing, not working." Perfect cover for cult: movement, crowds, visitors, noise, constant social events.',
 ARRAY['historical', 'diplomacy', 'political', 'vienna', '1814'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'British Delegation', 'organization',
 'British diplomatic mission at the Congress of Vienna, headquartered at Palais Modena.',
 'Lord Harcourt is an official observer. Wellington expected in autumn.',
 ARRAY['british', 'diplomatic', 'congress', 'vienna'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md');

-- ==========================================================================
-- ENTITIES — Key Locations
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Vienna', 'location',
 'Imperial capital of Austria. Setting for Chapter 3. A city teeming with diplomats, spies, and society intrigue during the Congress preparations.',
 'Surveillance state under Metternich''s Geheimpolizei. Strict court etiquette. The waltz is standard and expected. Affairs more openly acknowledged among aristocracy.',
 ARRAY['city', 'austria', 'chapter-3', 'imperial-capital'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'London', 'location',
 'Capital of Britain. Setting for Chapter 1. Where the investigators first encounter the Aeternum Choir through the Orphean Society.',
 NULL,
 ARRAY['city', 'britain', 'chapter-1'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Lyon', 'location',
 'Major French city. Setting for Chapter 2. Post-Revolutionary esotericism meets Gallo-Roman ruins.',
 'Politically volatile — Royalists vs. Bonapartists. The Bastille Day ritual takes place beneath Fourviere.',
 ARRAY['city', 'france', 'chapter-2'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Venice', 'location',
 'Italian city of canals. Setting for Chapter 2.5 (off-camera). The Confraternita del Bel Canto operated here.',
 'Resolved by Varrio Harrowmont. Not played at the table.',
 ARRAY['city', 'italy', 'chapter-2-5', 'off-camera'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Calcutta', 'location',
 'Capital of British India in 1814. Setting for Chapter 4. East India Company power base with British aristocratic expatriate society layered over Hindu ritual and river ceremony.',
 'Ritual exploits the Hooghly River''s acoustic properties and the monsoon break. Kali Puja / Diwali season.',
 ARRAY['city', 'india', 'chapter-4', 'british-india'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'University of Vienna', 'location',
 'Occupies the Jesuit College complex centred on Universitatsplatz. Medieval buildings rebuilt by Jesuits in the 17th century, now in administrative chaos following Jesuit suppression and restoration.',
 'Central quadrangle with arched galleries, lecture halls, library, faculty offices, and crucially the Medical Faculty Wing with basement cellars containing the sealed anatomical theatre.',
 ARRAY['building', 'vienna', 'university', 'chapter-3'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Sealed Anatomical Theatre', 'location',
 'Circular chamber beneath the University of Vienna, roughly 40 feet in diameter with 20-foot vaulted ceiling. Sealed in 1794, secretly reopened by Herzfeld for "acoustic research." Houses the Harmonic Engine.',
 'Whitewashed brick walls, elevated central platform, tiered stone benches for 40+ observers. Signs of recent activity: fresh lock scratches, candle stubs, chalk geometric patterns, new brass tubing and glass vessels. Access: main oak door (Hard lockpick), service corridor/crawlway, ventilation grates.',
 ARRAY['room', 'vienna', 'university', 'ritual-site', 'horror', 'chapter-3'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Gasthof zum Goldenen Adler', 'location',
 'The Golden Eagle Inn. Lodging option in the Innere Stadt near the Graben. 2-3 gulden per night.',
 'Central location, easy to blend in. Higher surveillance, informers among staff.',
 ARRAY['inn', 'vienna', 'lodging', 'innere-stadt'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Pension Hofmann', 'location',
 'Academic boarding house near the University in the Alsergrund district. Run by Frau Dorothea Hofmann. 1-2 gulden per night.',
 'Near University — good for investigation. Less respectable address, harder to access high society.',
 ARRAY['lodging', 'vienna', 'alsergrund', 'boarding-house'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Palais Modena', 'location',
 'Headquarters of the British delegation in Vienna during the Congress.',
 'Where Lord Harcourt can be found. Investigators can seek audience with junior diplomats.',
 ARRAY['building', 'vienna', 'british', 'diplomatic'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Hofburg Imperial Palace', 'location',
 'Imperial residence in Vienna. Site of the Imperial Reception on August 6 where the Harcourt reunion occurs.',
 'The pivotal campaign moment where Harcourt learns the full scope of the Choir''s global network.',
 ARRAY['building', 'vienna', 'imperial', 'palace'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Schonbrunn Palace', 'location',
 'Imperial summer palace. Site of the August 12 ball — the last major social event before the ritual.',
 'Very high danger level. Final victim acquisition event.',
 ARRAY['building', 'vienna', 'palace', 'ball-venue'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Palais Lobkowitz', 'location',
 'Site of the August 8 masquerade ball.',
 'High danger level. Opportunity to follow cult members under cover of masks.',
 ARRAY['building', 'vienna', 'palace', 'masquerade'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Burgtheater', 'location',
 'Imperial court theatre. Site of the August 10 concert, which doubles as a cult recruitment event.',
 NULL,
 ARRAY['building', 'vienna', 'theatre', 'concert-venue'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'St. Stephen''s Cathedral', 'location',
 'Vienna''s great cathedral. Public ceremonies on August 14 provide cover for the cult''s ritual preparations.',
 NULL,
 ARRAY['building', 'vienna', 'cathedral', 'church'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'The Crooked Chimney', 'location',
 'Der Schiefe Schornstein. Disreputable tavern in the Leopoldstadt district where Dr. Brenner hides, drinking in a back corner.',
 'Brenner treats injuries for locals who cannot afford real doctors. Cult assassins may attack here.',
 ARRAY['tavern', 'vienna', 'leopoldstadt', 'brenner-hideout'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Leopoldstadt', 'location',
 'District across the Danube Canal from central Vienna. Houses Vienna''s Jewish community and various marginal populations. Where Dr. Brenner is hiding.',
 NULL,
 ARRAY['district', 'vienna', 'jewish-quarter'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Naturhistorisches Institut', 'location',
 'The Natural History Institute in Vienna where Dr. Ernst Falkner works as a geologist.',
 'Investigators'' first contact point in Vienna. Falkner can be reached here using the code phrase.',
 ARRAY['building', 'vienna', 'science', 'falkner'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Stonehenge', 'location',
 'Ancient stone circle in England. Ritual site for the London cell (Orphean Society). Where Segment I — Core Harmonic Framework — was performed.',
 'Ritual destroyed by the party in June 1814.',
 ARRAY['ritual-site', 'britain', 'chapter-1', 'ancient'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Roman Amphitheatre beneath Fourviere', 'location',
 'Roman ruins beneath the Fourviere hill in Lyon. Ritual site for the Lyon cell. Acoustic properties serve as a resonance chamber for the Segment II ritual.',
 'Ritual on July 14 (Bastille Day).',
 ARRAY['ritual-site', 'lyon', 'chapter-2', 'roman', 'underground'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md'),

(36, 'Hotel de Brissac', 'location',
 'Venue for the July 3rd Lyon Ball hosted by Duchesse Camille de Brissac.',
 NULL,
 ARRAY['building', 'lyon', 'ball-venue', 'chapter-2'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt'),

(36, 'Linienwall Gates', 'location',
 'Vienna''s customs checkpoint where arriving travellers are inspected. Site of the investigators'' entry into Vienna on August 3, 1814.',
 'Madame Delacroix encountered here. Inspector Novak conducts thorough inspections.',
 ARRAY['vienna', 'checkpoint', 'customs', 'arrival-scene'],
 'AUTHORITATIVE', 'Vienna_Supplemental_Notes.md'),

(36, 'Munich', 'location',
 'City in Bavaria. Base of operations for "Der Kantor," the Aeternum Choir''s Central European coordinator.',
 'Coded correspondence between cells routes through Der Kantor.',
 ARRAY['city', 'bavaria', 'choir-coordination'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md');

-- ==========================================================================
-- ENTITIES — Key Items & Artifacts
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'The Harmonic Engine', 'artifact',
 'A monstrous 30-foot biomechanical instrument built from human body parts. Bronzed esophageal tubes as pipes, human lungs as bellows, arms and hands wired to key levers, and a living human brain at its heart interpreting the Canticle score. Victims are conscious but paralysed via alchemical neurotoxins.',
 'Located in the sealed anatomical theatre beneath the University. Self-playing performance of Segment V at midnight August 15. 6-8 individuals in various states of dismemberment. SAN Loss: 1d6/1d10 to witness. Design note: the ritual MUST succeed (see Campaign_Design_Notes.md). The Engine may have protections: wards, chemical traps, dimensional bleed causing temporal/spatial distortions.',
 ARRAY['horror', 'biomechanical', 'ritual-device', 'organ', 'vienna', 'brotherhood'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Herzfeld''s Letter to Savarin', 'document',
 'Letter from Professor Herzfeld to Mathilde Savarin, recovered from Lyon. Mentions Vienna, August 15, and the "Brotherhood of the Open Measure."',
 'Key clue that directed investigators to Vienna. May cause complications at customs if found by Inspector Novak.',
 ARRAY['clue', 'letter', 'handout', 'vienna', 'lyon'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'The Grand Canticle Score', 'artifact',
 'The complete notation for the Grand Canticle of Yog-Sothoth. Each segment is encoded in harmonic cipher — mathematically perfect arrangements designed to resonate with Yog-Sothoth''s eternal presence.',
 'Herzfeld''s complete notation for Segment V can be found in his University office or apartments. If captured, reveals the mathematical structure underlying all the rituals.',
 ARRAY['ritual', 'music', 'cipher', 'mythos', 'clue'],
 'AUTHORITATIVE', 'Campaign_Overview_Updated.md');

-- ==========================================================================
-- ENTITIES — Key Events
-- ==========================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, tags, source_confidence, source_document) VALUES
(36, 'Imperial Reception', 'event',
 'Formal reception at the Hofburg on August 6, 1814. The pivotal campaign moment where the investigators unexpectedly reunite with Lord Harcourt and Lady Honoria.',
 'Neither group knows the other is in Vienna until this moment. After this reunion, Harcourt learns the full scope of the Choir''s network and begins dispatching Order agents. Campaign shifts from ad hoc investigation to strategic coordination.',
 ARRAY['vienna', 'social-event', 'reunion', 'chapter-3', 'pivotal'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Masquerade at Palais Lobkowitz', 'event',
 'Masquerade ball on August 8, 1814. Opportunity to follow cult members under cover of masks.',
 'High danger level.',
 ARRAY['vienna', 'social-event', 'masquerade', 'chapter-3'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'Salon at Countess von Thun', 'event',
 'Social salon hosted by Countess von Thun on August 5, 1814. The investigators'' introduction to Viennese high society.',
 'Low danger level. Must secure an invitation first — ten possible paths detailed in supplemental notes.',
 ARRAY['vienna', 'social-event', 'salon', 'chapter-3'],
 'AUTHORITATIVE', 'Vienna_Chapter_Summary.md'),

(36, 'The August 15 Ritual', 'ritual',
 'The Harmonic Engine performs Segment V of the Grand Canticle at midnight on August 15, 1814 (Feast of the Assumption). The Brotherhood''s culminating ceremony.',
 'DESIGN DECISION: This ritual MUST succeed. The PCs must try desperately to stop it and fail. The failure must feel earned, not railroaded. Multiple failure vectors needed. See Campaign_Design_Notes.md for full requirements.',
 ARRAY['ritual', 'vienna', 'chapter-3', 'climax', 'must-succeed'],
 'AUTHORITATIVE', 'Campaign_Design_Notes.md'),

(36, 'Lyon Ball', 'event',
 'Grand ball at Hotel de Brissac on July 3, 1814. The investigators attend and encounter Savarin''s influence, the Blue Sash enforcers, and experience acoustic phenomena.',
 'Eight distinct scenes from 8PM to midnight. A servant collapses from resonance trauma. Savarin tests harmonic effects through live music. The entire ball is a rehearsal.',
 ARRAY['lyon', 'social-event', 'ball', 'chapter-2'],
 'AUTHORITATIVE', 'Lyon Ball Details.txt');

-- ==========================================================================
-- RELATIONSHIPS
-- ==========================================================================

-- Helper: look up entity IDs by name within campaign 36
-- Relationship type IDs from the relationship_types table

-- ---- Order of St. Aelfric memberships ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Senior operative and leader'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lord Percival Harcourt'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'Leads the Order as Earl of Wrexham'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lord Percival Harcourt'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Member and investigator trainer'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lady Honoria Lyndhurst'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Employed as investigators by the Order'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Marina Garrick'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Employed as investigators by the Order'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Emma Wentworth'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Employed as investigators by the Order'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Georgiana Wentworth'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Employed as investigators by the Order'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Charlotte Thorne'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Employed as investigators by the Order'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Colonel Henri Moreau'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Full member — disrupted the Venice cell independently'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Varrio Harrowmont'
AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

-- ---- Harcourt / Honoria relationship ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 803, 'Companion and travelling partner'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lord Percival Harcourt'
AND t.campaign_id = 36 AND t.name = 'Lady Honoria Lyndhurst';

-- ---- Honoria mentored the investigators ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 819, 'Trained the investigators in London'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lady Honoria Lyndhurst'
AND t.campaign_id = 36 AND t.name = 'Marina Garrick';

-- ---- Wentworth sisters ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 845, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Emma Wentworth'
AND t.campaign_id = 36 AND t.name = 'Georgiana Wentworth';

-- ---- Brotherhood memberships ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'Cult leader and Engine designer'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Professor Albin Herzfeld'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Lieutenant and recruiter'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Kapellmeister Friedrich Adler'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 839, 'Reports to Herzfeld as second-in-command'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Kapellmeister Friedrich Adler'
AND t.campaign_id = 36 AND t.name = 'Professor Albin Herzfeld';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 799, 'Financial backer providing funding and political protection'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Count Leopold von Trauttmansdorff'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Secret sympathizer, not full member'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Baron Otto von Kaunitz'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'True believer and Geheimpolizei mole'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Inspektor Heinrich Vogel'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 812, 'Mole infiltrating the police for the cult'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Inspektor Heinrich Vogel'
AND t.campaign_id = 36 AND t.name = 'Geheimpolizei';

-- ---- Brotherhood / Aeternum Choir ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Vienna cell of the global cult'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Brotherhood of the Open Measure'
AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Lyon cell of the global cult'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Societe Harmonique de l''Aube'
AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Venice cell of the global cult (destroyed)'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Confraternita del Bel Canto'
AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'London cell of the global cult (destroyed)'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Orphean Society'
AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

-- ---- Cult worships Yog-Sothoth ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 861, 'Seeks cosmic transcendence through the Harmonic Totality'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Aeternum Choir'
AND t.campaign_id = 36 AND t.name = 'Yog-Sothoth';

-- ---- Lyon cult relationships ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'High priestess and conductor of ritual'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mathilde Savarin'
AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 844, 'Commander of the Blue Sash enforcers serving the cult'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Captain Luc Fouchard'
AND t.campaign_id = 36 AND t.name = 'Mathilde Savarin';

-- ---- Brenner / Brotherhood (defection) ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 790, 'Former member who defected after participating in early Engine experiments'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Wilhelm Brenner'
AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 808, 'Tasked with silencing Brenner permanently'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Kapellmeister Friedrich Adler'
AND t.campaign_id = 36 AND t.name = 'Dr. Wilhelm Brenner';

-- ---- Falkner''s daughter ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 827, 'Father — daughter vanished 3 years ago to a "musical academy"'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Ernst Falkner'
AND t.campaign_id = 36 AND t.name = 'Margarethe Falkner';

-- ---- Engine victims ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Soprano component integrated into the Engine'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Anna Lindqvist'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'May be integrated as soprano component'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Margarethe Falkner'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 787, 'Designed and built the Harmonic Engine'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Professor Albin Herzfeld'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

-- ---- Herzfeld / University ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Professor at the University'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Professor Albin Herzfeld'
AND t.campaign_id = 36 AND t.name = 'University of Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Located beneath the University in the sealed theatre'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'The Harmonic Engine'
AND t.campaign_id = 36 AND t.name = 'Sealed Anatomical Theatre';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Part of the University complex'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Sealed Anatomical Theatre'
AND t.campaign_id = 36 AND t.name = 'University of Vienna';

-- ---- Falkner location ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Works at the Naturhistorisches Institut'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Ernst Falkner'
AND t.campaign_id = 36 AND t.name = 'Naturhistorisches Institut';

-- ---- Brenner location ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Hides at the Crooked Chimney tavern'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Wilhelm Brenner'
AND t.campaign_id = 36 AND t.name = 'The Crooked Chimney';

-- ---- Harcourt at Palais Modena ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Staying at the British delegation headquarters'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lord Percival Harcourt'
AND t.campaign_id = 36 AND t.name = 'Palais Modena';

-- ---- Metternich / Geheimpolizei ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 784, 'Runs Vienna''s secret police'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Prince Klemens von Metternich'
AND t.campaign_id = 36 AND t.name = 'Geheimpolizei';

-- ---- Von Thun / niece ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, t.id, s.id, 827, 'Aunt to Elise'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Fraulein Elise von Schonberg'
AND t.campaign_id = 36 AND t.name = 'Countess Maria von Thun';

-- ---- Delacroix / von Thun connection ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 814, 'Has letter of introduction — late husband''s godmother'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Madame Celestine Delacroix'
AND t.campaign_id = 36 AND t.name = 'Countess Maria von Thun';

-- ---- Sternberg / Wyndham rivalry ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 841, 'Romantic rival — may provoke a duel'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Graf Maximilian von Sternberg'
AND t.campaign_id = 36 AND t.name = 'Thomas Wyndham';

-- ---- Varrio disrupted Venice ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 824, 'Disrupted the Venice cell from within'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Varrio Harrowmont'
AND t.campaign_id = 36 AND t.name = 'Confraternita del Bel Canto';

-- ---- Kaunitz reports to Herzfeld ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 839, 'Reports investigators'' movements to Herzfeld'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Baron Otto von Kaunitz'
AND t.campaign_id = 36 AND t.name = 'Professor Albin Herzfeld';

-- ---- Adler''s sister in the Engine ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 781, 'His sister''s voice is part of the Engine'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Kapellmeister Friedrich Adler'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

-- ---- Holzer and Gruber / Engine ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Donated left lung and voice box to the Engine'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Frau Margarethe Holzer'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Donated both hands to the Engine'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Hans Gruber'
AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

-- ---- Location containment ----
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'University of Vienna'
AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Hofburg Imperial Palace'
AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Gasthof zum Goldenen Adler'
AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'The Crooked Chimney'
AND t.campaign_id = 36 AND t.name = 'Leopoldstadt';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Leopoldstadt'
AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Roman Amphitheatre beneath Fourviere'
AND t.campaign_id = 36 AND t.name = 'Lyon';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Hotel de Brissac'
AND t.campaign_id = 36 AND t.name = 'Lyon';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, NULL
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Stonehenge'
AND t.campaign_id = 36 AND t.name = 'London';

COMMIT;

-- ==========================================================================
-- Summary
-- ==========================================================================
-- Chapters: 5
-- Entities: ~75 (6 PCs, ~35 NPCs, 1 deity, ~10 organizations/cults,
--                ~20 locations, 3 items/artifacts, ~5 events/rituals)
-- Relationships: ~50
-- ==========================================================================
