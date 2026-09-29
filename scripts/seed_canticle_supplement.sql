-- ============================================================================
-- Canticle Of The End — Supplemental Seed Data
-- ============================================================================
-- Campaign ID: 36
-- This supplement adds entities extracted from session-level planning docs,
-- play notes, analysis files, and chapter source material not covered in the
-- initial seed (seed_canticle.sql).
--
-- Run with:
--   docker exec -i imagineer-postgres psql -U imagineer -d imagineer \
--     < scripts/seed_canticle_supplement.sql
-- ============================================================================

BEGIN;

-- ============================================================================
-- PCs
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Augustus Bolt', 'pc',
 'Explorer, map-maker, and rationalist. Order of St. Aelfric recruit who joined the investigation in London.',
 'Deceased. Killed during the raid on the Orphans Hospital in Lyon (Chapter 2). Replaced in the party by Colonel Henri Moreau.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'chapter-2']),

(36, 'Adrien, Viscount de Montferrand', 'pc',
 'French aristocrat, former cavalry officer turned naturalist. Diplomatic charmer who defused the Austrian officer confrontation at the opera and danced with Caroline Hartley at the Imperial Reception.',
 'Active PC as of Session 4. Owns a hunting lodge above Lyon with an occult library. Introduced by Comte de Puyrault in Lyon. Received a Major Wound during the hospital raid. Romantic rapport developing with Caroline Hartley.',
 'AUTHORITATIVE', ARRAY['active', 'chapter-2', 'chapter-3']),

(36, 'Jacob', 'pc',
 'Investigator who accompanied the party through London and Lyon.',
 'Deceased. Killed during the raid on the Orphans Hospital in Lyon (Chapter 2). Minimal background detail in source files.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'chapter-2']);

-- ============================================================================
-- NPCs — London (Chapter 1)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Dr. Julius Hume', 'npc',
 'Ritualist and high cantor of the London Orphean Society. Conducted the Stonehenge ritual attempting to consecrate Segment I of the Grand Canticle.',
 'Deceased. Killed at Stonehenge when the investigators disrupted the ritual. Correspondent of Mathilde Savarin and Professor Herzfeld. Herzfeld''s letter to Hume was found at the Orphean Society offices.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'orphean-society']),

(36, 'Lady Aurelia Danforth', 'npc',
 'Second-in-command of the Orphean Society. Soprano who composed unsanctioned Canticle variations and took liberties with the notation, drawing criticism from Herzfeld.',
 'Deceased. Died at Stonehenge during the Chapter 1 ritual. Referenced as "Lady D." in Herzfeld''s letter. Had ambitions to control the Canticle independently.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'orphean-society']);

-- ============================================================================
-- NPCs — Lyon (Chapter 2)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Comte Emeric de Puyrault', 'npc',
 'Lyonnais aristocrat, age 36. Master fencer and Order of St. Aelfric sympathiser. Social sponsor of the investigators in Lyon.',
 'Alive. Romantic interest of Georgiana Wentworth. Brother of Etienne de Puyrault (missing 6 months, presumed dead). Owns Le Coteau des Ombres estate on Lyon''s northern outskirts.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2', 'order-of-st-aelfric']),

(36, 'Sabine Lambert', 'npc',
 'Priestess of the Lyre within the Societe Harmonique de l''Aube. Oversees musical conditioning of child victims and ritual music preparation.',
 'Alive. Inner circle cultist in Lyon. Distinct from Annette Lambert (music librarian/archivist, cult sympathiser).',
 'AUTHORITATIVE', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Claudine Morel', 'npc',
 'Watcher and infiltrator for the Societe Harmonique. Known as the "Lady in Grey" at the Paris Ball. Expert at shapeshifting social class.',
 'Alive. Shadows the investigators and reports to Savarin.',
 'AUTHORITATIVE', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Victoire Lenoir', 'npc',
 'Cult soprano and Siren of the Choir. Narcissistic performer who sang for Savarin directly and confronted Emma Wentworth at the soiree.',
 'Alive. Also listed as Victoire Lefevre in some sources — canonical name needs GM confirmation.',
 'DRAFT', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Lucien Goupil', 'npc',
 'Nouveau riche social climber and Jacobin dilettante. Hosted the July 1st masked soiree at Maison du Corbeau where the Formless Spawn was summoned.',
 'Alive. Cult sympathiser and willing dupe seeking status with Savarin. Not a true believer.',
 'AUTHORITATIVE', ARRAY['cult-sympathiser', 'chapter-2']),

(36, 'Magistrate Beraud', 'npc',
 'Municipal magistrate and Commissaire Delaroche''s superior. Bribed by Savarin and effectively in her pocket.',
 'Alive. Savarin''s political shield in Lyon. Delaroche must report to him.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-2', 'corrupted']),

(36, 'Vicomte Henri Lavigne', 'npc',
 'Impoverished music patron and antiquarian ruined by Savarin. Attended the Paris Ball at Hotel de Brissac.',
 'Alive. Shared his history with Charlotte Thorne. Knows about the Orphic Lyre from personal experience.',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-2']),

(36, 'Contessa Erminia Foscarini', 'npc',
 'Withered Venetian noblewoman and conductor of the Venice Stillwater Choir ritual. Held the corrupted Orphic Lyre.',
 'Deceased. Killed when the Venice ritual collapsed after Harrowmont''s intervention. Leader of the Confraternita dell''Acqua Ferma.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-2', 'venice-cult']),

(36, 'Madame Leontine des Jardins', 'npc',
 'Wealthy widow and patroness of the Societe Harmonique. Host of the Maison du Choeur Sacre soiree.',
 'Alive. Outer-circle facilitator for Savarin. Not fully aware of the cult''s true scope.',
 'AUTHORITATIVE', ARRAY['cult-sympathiser', 'chapter-2']);

-- ============================================================================
-- NPCs — Vienna (Chapter 3)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Lady Ashworth', 'npc',
 'English dowager in Vienna. Social guide and chaperone who provides carriage access, opera box seating, and introductions to Viennese high society.',
 'Alive. Patron/protector of the PCs. Connected to Countess von Thun. Warned Emma about Kaunitz at the opera.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-3']),

(36, 'Freddie Cavendish', 'npc',
 'Wealthy, clueless young English socialite. Secured masquerade invitations for the party. Travels with his manservant Pemberton.',
 'Alive. Has a drinks cabinet in his coach. Connected to the Palais Kinsky social circle.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-3']),

(36, 'Klaus Bauer', 'npc',
 'Former soldier and Brotherhood enforcer. Methodical and professional. Procures bodies for the Harmonic Engine and guards the sealed anatomical theatre.',
 'Alive as of Session 4. Twin of Werner Bauer. Carries a university-badged key. Reassigned to assassination duty targeting the investigators in Session 5.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood']),

(36, 'Werner Bauer', 'npc',
 'Former soldier and Brotherhood enforcer. Aggressive and direct. Works alongside his twin Klaus procuring bodies and guarding the Engine.',
 'Alive as of Session 4. Twin of Klaus Bauer. Reassigned to assassination duty in Session 5.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood']),

(36, 'Widow Katz', 'npc',
 'Sharp-eyed Jewish landlady in Leopoldstadt who housed Dr. Brenner at her boarding house with the yellow door.',
 'Alive. Saw the party visit Brenner the night of his death. Gave descriptions to Adler. Sent a warning note to Varrio via a street urchin. Will be interrogated by cult operatives.',
 'AUTHORITATIVE', ARRAY['witness', 'chapter-3']),

(36, 'Mr. Hartley', 'npc',
 'Wealthy English merchant in Vienna seeking investment opportunities. Dangerously fawns over aristocratic connections and is unwittingly approaching cult members about industrial partnerships.',
 'Alive. Father of Caroline and Lydia Hartley. Spoke with Trauttmansdorff (dangerous). Potential unwitting cult victim.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Caroline Hartley', 'npc',
 'Intelligent, cultured English woman who enjoys music, Roman history, and art. Developing a genuine rapport with Adrien after dancing at the Imperial Reception.',
 'Alive. Daughter of Mr. Hartley. Romantic interest for PC Adrien. Potential confidant.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3', 'romantic-interest']),

(36, 'Lydia Hartley', 'npc',
 'Younger daughter of Mr. Hartley. Infatuated with Adrien, who treats her like a younger sister.',
 'Alive. Potential cult victim if Herzfeld takes interest in the family.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Major Konstantin Thurner', 'npc',
 'Retired Austrian officer serving as Harcourt''s trusted safe house contact in Vienna. Provides weapons cache, clean clothes, and medical supplies from Josefstadt district.',
 'Alive. Operates safe house at Lange Gasse 14 in Josefstadt. Order of St. Aelfric logistics support.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-3', 'order-of-st-aelfric']),

(36, 'Sergeant Franz Huber', 'npc',
 'Street-level Geheimpolizei agent bribed by Kaunitz to report on persons of interest. Purely mercenary, not a true believer. Scarred cheek.',
 'Alive. Reports to Kaunitz at Cafe Frauenhuber. Part of the corrupted police network.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'corrupted']),

(36, 'Greta Adler', 'npc',
 'Friedrich Adler''s sister. One of the early "volunteers" whose voice box is now integrated into the Harmonic Engine''s soprano register.',
 'Deceased / integrated into the Engine. Her fate is Adler''s psychological weakness and the source of his guilt-driven fanaticism.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-3', 'engine-victim']),

(36, 'Dr. Gottfried Reiner', 'npc',
 'University chemist who produces neurotoxins and preservation fluids for the Harmonic Engine, believing he is developing revolutionary anaesthetics.',
 'Alive. Brotherhood outer circle — does not know the Engine''s true nature. Reports to Herzfeld.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood', 'unwitting']),

(36, 'Herr Gustav Metzger', 'npc',
 'Master instrument maker building mechanical components for the Harmonic Engine, unaware of its biological elements.',
 'Alive. Brotherhood outer circle — genuinely believes he is building a revolutionary musical instrument.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood', 'unwitting']),

(36, 'Fraulein Liesel Hartmann', 'npc',
 'Nurse who manages victim care and sedation for the Engine''s "volunteers," believing she works in an experimental medical programme.',
 'Alive. Brotherhood outer circle. Administers sedatives to victims before integration.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood', 'unwitting']),

(36, 'Frau Ingrid Sperl', 'npc',
 'Housekeeper who maintains the sealed anatomical theatre, believing it is a private research laboratory. Terrified of Adler.',
 'Alive. Brotherhood outer circle. Her fear of Adler could be exploited by investigators.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'brotherhood', 'unwitting']),

(36, 'Der Kantor', 'npc',
 'Codename for the Aeternum Choir''s Central European coordinator, based in Munich. Communicates with cells via coded letters.',
 'Alive. Coordinates between all European cells. Sent the August 9 warning letter to Herzfeld about investigator interference. True identity unknown.',
 'AUTHORITATIVE', ARRAY['antagonist', 'campaign-wide', 'aeternum-choir']),

(36, 'Henri Valmont', 'npc',
 'French violinist in Vienna who reported disturbing rumours about auditions that lead nowhere — musicians go in but don''t come out.',
 'Alive. Independent source on musician disappearances. Contact of Count de la Tour-du-Pin.',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-3']),

(36, 'Herr Braun', 'npc',
 'Vogel''s surveillance agent. The persistent man in the brown coat watching Palais Kinsky from Am Hof Square.',
 'Alive. Reports everything to Inspektor Vogel. Not personally dangerous.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-3', 'geheimpolizei']);

-- ============================================================================
-- Missing Russian Musicians (Vienna victims)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Dmitri Sokolov', 'npc',
 'Russian cellist, age 34. Last seen July 28 near the Graben.',
 'Unknown status — likely dead, integrated into the Harmonic Engine. Listed in Volkonsky''s notebook of missing Russian musicians.',
 'AUTHORITATIVE', ARRAY['missing', 'chapter-3', 'engine-victim']),

(36, 'Alexei Petrov', 'npc',
 'Russian violinist, age 26. Last seen July 31.',
 'Unknown status — likely dead, integrated into the Harmonic Engine. Listed in Volkonsky''s notebook.',
 'AUTHORITATIVE', ARRAY['missing', 'chapter-3', 'engine-victim']),

(36, 'Ivan Markov', 'npc',
 'Russian French horn player, age 29. Last seen August 2.',
 'Unknown status — likely dead, integrated into the Harmonic Engine. Listed in Volkonsky''s notebook.',
 'AUTHORITATIVE', ARRAY['missing', 'chapter-3', 'engine-victim']),

(36, 'Sergei Orlov', 'npc',
 'Russian viola player, age 31. Last seen August 4.',
 'Unknown status — likely dead, integrated into the Harmonic Engine. Listed in Volkonsky''s notebook.',
 'AUTHORITATIVE', ARRAY['missing', 'chapter-3', 'engine-victim']);

-- ============================================================================
-- Creatures
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Chakota', 'creature',
 'Mythos entity associated with Aeternox''s reflection aspect. Composed of absorbed children, mouths, and memories. Warped through water and resonance.',
 'Kept in the basement of the Silkweavers Guild in Lyon behind a guarded wooden screen, fed excessively to keep it stable. Controlled by the Orphic Lyre. Status after Lyon cell destruction uncertain.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Ciimba', 'creature',
 'Undead/zombified children used as guards and ritual components by the Societe Harmonique. Created through Doctor Carreau''s surgical experiments.',
 'Active during Lyon chapter, guarding the practice chamber beneath the Silkweavers Guild. Status after Lyon cell destruction uncertain.',
 'AUTHORITATIVE', ARRAY['mythos', 'undead', 'chapter-2']),

(36, 'Formless Spawn', 'creature',
 'Servitor of Tsathoggua summoned during the July 1st masked soiree at Maison du Corbeau. Required blood offering and child sacrifice.',
 'Dismissed after being sated by sacrifice. The summoning was the secret ritual behind the soiree.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Nightgaunt', 'creature',
 'Faceless, winged Mythos entity drawn to Vienna by the Harmonic Engine''s dimensional bleed. Attacks with a paralysing barbed tail.',
 'To be encountered in Session 5. A "harmonic sending" that follows harmonic residue on PCs who heard the Beethoven moment at the Imperial Reception. Attacks after midnight on August 7-8.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-3']),

(36, 'Awakened Drowned', 'creature',
 'Waterlogged revenants raised from Venice''s canals during the Stillwater Choir ritual. Animated by the Canticle''s resonance.',
 'Collapsed when the Venice ritual was broken by Harrowmont, but Venice is described as "no longer truly asleep." Dormant.',
 'AUTHORITATIVE', ARRAY['mythos', 'undead', 'chapter-2']);

-- ============================================================================
-- Deities
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Aeternox', 'deity',
 'Fictionalised Roman deity of eternity, darkness, deep time, and subterranean rites. Worshipped in underground chapels in ancient Lugdunum. Associated with echo, resonance, and thresholds.',
 'Conceptual framework for the Lyon cult. The Temple of Aeternox beneath Fourviere was the site of Savarin''s planned Bastille Day ritual. Aliases: Aeternox Invictus, Aeternox Subterranus, Aeternox Vocis.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Tsathoggua', 'deity',
 'Great Old One. The Formless Spawn summoned at the July 1st soiree is a servitor of Tsathoggua.',
 'Referenced but not directly encountered. The summoning at the soiree was a secondary ritual unrelated to the Grand Canticle.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Kali', 'deity',
 'Hindu goddess of time, death, transformation, and destruction. Kali Puja on October 25, 1814 provides the spiritual cover and psychic resonance for the Calcutta ritual.',
 'The Kali Puja/Diwali overlap on the new moon creates the ritual window for the Calcutta cell. Not a Mythos entity per se — her worship is exploited by the cult.',
 'AUTHORITATIVE', ARRAY['chapter-4']);

-- ============================================================================
-- Locations — London
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Kensington Stables', 'location',
 'Nondescript yard in west London marked on Hume''s map as a site of hidden Orphean Society activity.',
 'Chapter 1 location.',
 'AUTHORITATIVE', ARRAY['chapter-1']);

-- ============================================================================
-- Locations — Lyon
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Paris', 'location',
 'Capital of France. The investigators attended the Ball at Hotel de Brissac before travelling south to Lyon.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Le Coteau des Ombres', 'location',
 'The Puyrault estate on Lyon''s northern outskirts beyond Croix-Rousse hill. Investigators'' base of operations during the Lyon chapter.',
 'Owned by Comte Emeric de Puyrault.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Silkweavers Guild', 'location',
 'Abandoned guild building in Lyon''s Croix-Rousse district. Headquarters of the Societe Harmonique de l''Aube, with a basement housing the practice choir, the Chakota, and Ciimba guards.',
 'Location where Herzfeld''s letter to Savarin was discovered. Connected to tunnel system beneath Lyon.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'cult-site']),

(36, 'Orphans Hospital', 'location',
 'Hospital in Lyon run by Doctor Carreau. Site of children''s imprisonment, experimentation, and the investigators'' raid that killed Augustus Bolt and Jacob.',
 'Over 240 children rescued during the raid.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Temple of Aeternox', 'location',
 'Underground excavated Roman temple beneath Fourviere Hill. Site of Savarin''s planned Bastille Day ritual to open "Aeternox''s Gate of Duration."',
 'Connected to the Silkweavers Guild tunnels.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'cult-site']),

(36, 'Maison du Corbeau', 'location',
 'Goupil''s inherited estate in Lyon''s old ecclesiastical district. Venue for the July 1st masked soiree where the Formless Spawn was summoned.',
 'Owned by Lucien Goupil.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Fourviere Hill', 'location',
 'Ancient hill in Lyon crowned with Roman ruins. Contains the excavated Temple of Aeternox beneath its surface.',
 'The Roman Amphitheatre beneath Fourviere (already in DB) is a separate but nearby location.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'San Trionfo', 'location',
 'Half-flooded basilica ruins beneath Venice. Site of the Stillwater Choir ritual that Varrio Harrowmont disrupted.',
 'Venice cult site. The ritual raised the Awakened Drowned from the canals.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'cult-site']);

-- ============================================================================
-- Locations — Vienna
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Palais Kinsky', 'location',
 'Elegant residence on Am Hof Square where the investigators are lodging in Vienna. Rooms were professionally searched by the Polizeidirektion on August 6.',
 'Party''s base of operations. Marina''s notebook and Mythos tomes stolen during the room search. Kaunitz''s white rose left on Emma''s dressing table.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Cafe Frauenhuber', 'location',
 'Coffee house on Himmelpfortgasse where Baron von Kaunitz meets his Geheimpolizei contacts every afternoon between 2-3 PM.',
 'Kaunitz meets Sergeant Huber here. Brotherhood wax seal (tuning fork in spiral) observed on correspondence.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Josefstadt District', 'location',
 'Respectable but not fashionable Viennese district. Location of Madame Delacroix''s pension and Major Thurner''s safe house at Lange Gasse 14.',
 'Contains weapons cache, medical supplies, and fallback position for the investigators.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Vienna Conservatory', 'location',
 'Musical institution where Kapellmeister Adler teaches, scouts talent, and cultivates Anna Lindqvist as the soprano target for the Engine.',
 'Adler''s daytime base. Anna Lindqvist practises here.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Eroica Hall', 'location',
 'Hall within Palais Lobkowitz where Beethoven premiered his Third Symphony. Site of the Grand Masquerade planned for August 8.',
 'The masquerade is a major social set-piece for Session 5.',
 'AUTHORITATIVE', ARRAY['chapter-3']);

-- ============================================================================
-- Locations — Calcutta
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Kalighat Temple', 'location',
 'Centre of Kali worship in all of Bengal and spiritual heart of Calcutta. Potential site near the Calcutta ritual.',
 'Primary Kali Puja celebration site. Chapter 4.',
 'AUTHORITATIVE', ARRAY['chapter-4']),

(36, 'Hooghly River', 'location',
 'River running through Calcutta whose acoustic properties and cremation ghats are exploited by the cult for funerary chant resonance.',
 'Referenced in Herzfeld''s letter as "where the Hugli bends inland." Chapter 4.',
 'AUTHORITATIVE', ARRAY['chapter-4']),

(36, 'Fort William', 'location',
 'British military stronghold in Calcutta and seat of the Governor-General''s administration in 1814.',
 'Centre of British military and political power in India. Chapter 4.',
 'AUTHORITATIVE', ARRAY['chapter-4']),

(36, 'Trieste', 'location',
 'Austrian port city serving as the departure point for the sea journey from Vienna to Calcutta via the Mediterranean.',
 'Travel waypoint between chapters 3 and 4.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'chapter-4']);

-- ============================================================================
-- Artifacts
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'The Orphic Lyre', 'artifact',
 'Corrupted ancient lyre used by Savarin to control the Chakota and direct the Lyon ritual''s harmonic resonance. Previously held by Contessa Foscarini in Venice.',
 'Key ritual instrument. Status after Lyon cell destruction uncertain.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Liber Ivonis', 'artifact',
 'Mythos tome stolen from the investigators'' rooms at Palais Kinsky during the August 6 police search.',
 'Now in cult possession. Part of the investigators'' occult library built up across London and Lyon.',
 'AUTHORITATIVE', ARRAY['mythos', 'stolen', 'chapter-3']),

(36, 'Cultes des Goules', 'artifact',
 'Mythos tome stolen from the investigators'' rooms at Palais Kinsky during the August 6 police search.',
 'Now in cult possession.',
 'AUTHORITATIVE', ARRAY['mythos', 'stolen', 'chapter-3']),

(36, 'De Vermis Mysteriis', 'artifact',
 'Mythos tome stolen from the investigators'' rooms at Palais Kinsky during the August 6 police search.',
 'Now in cult possession.',
 'AUTHORITATIVE', ARRAY['mythos', 'stolen', 'chapter-3']),

(36, 'Revelations of Glaaki', 'artifact',
 'Mythos tome stolen from the investigators'' rooms at Palais Kinsky during the August 6 police search.',
 'Now in cult possession.',
 'AUTHORITATIVE', ARRAY['mythos', 'stolen', 'chapter-3']),

(36, 'Carreau''s Ledger of Children', 'artifact',
 'Notebook documenting kidnapped children with names, ages, family connections, and "Harmonic Susceptibility Index." Major political evidence recovered during the Orphans Hospital raid.',
 'In investigators'' possession. Critical evidence linking the cult to child abduction.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-2']);

-- ============================================================================
-- Documents
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Marina''s Notebook', 'document',
 'Complete operational history of the investigators'' campaign activities across London, Lyon, and Venice. Stolen from Palais Kinsky during the August 6 room search.',
 'Now in Herzfeld''s hands. Catastrophic intelligence loss — reveals the party destroyed London and Lyon cells, their methods, capabilities, and Order of St. Aelfric connections.',
 'AUTHORITATIVE', ARRAY['stolen', 'chapter-3']),

(36, 'Herzfeld''s Letter to Hume', 'document',
 'Letter from Herzfeld to Dr. Hume on University of Vienna cream stationery. Criticises tempo fluctuations in Segment IV and Lady Danforth''s unsanctioned notation changes.',
 'Found at Orphean Society offices in London. Establishes the Herzfeld-Hume correspondence and links Vienna to London.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-1']),

(36, 'Dorothea''s Intelligence Dossier', 'document',
 'Written summary in French detailing operational intelligence on all Brotherhood cult members'' schedules, habits, and vulnerabilities.',
 'In investigators'' possession. Provided by Dorothea de Courlande at her salon. Covers Herzfeld, Adler, Kaunitz, Trauttmansdorff, and police infiltration.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-3']),

(36, 'Volkonsky''s Notebook', 'document',
 'Small leather notebook containing names, dates, and last known locations of four missing Russian musicians (Sokolov, Petrov, Markov, Orlov).',
 'In Georgiana''s possession. Given by Count Volkonsky at the Imperial Reception.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-3']);

-- ============================================================================
-- Events
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Stonehenge Ritual', 'event',
 'The Orphean Society''s ritual at Stonehenge on June 12, 1814 to consecrate Segment I of the Grand Canticle. Disrupted by the investigators with heavy casualties.',
 'Dr. Hume and Lady Danforth died. Cell destroyed. Chapter 1 climax.',
 'AUTHORITATIVE', ARRAY['chapter-1', 'ritual-disrupted']),

(36, 'Orphans Hospital Raid', 'event',
 'Raid by investigators, Colonel Moreau, and veterans to rescue 240+ children from Doctor Carreau''s hospital in Lyon.',
 'Augustus Bolt and Jacob killed. Doctor Carreau killed. Over 240 children rescued. Chapter 2 climax.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'July 1st Masked Soiree', 'event',
 'Masked musical soiree at Maison du Corbeau hosted by Lucien Goupil. Investigators'' first encounter with cult ritual — a Formless Spawn was summoned via blood offering.',
 'Key Chapter 2 event. Emma confronted Victoire Lenoir. Georgiana detected Mythos undertones in the music.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Palais Kinsky Room Search', 'event',
 'Professional search of the investigators'' rooms on August 6 by the Polizeidirektion while the party attended the Imperial Reception.',
 'All Mythos tomes and Marina''s notebook stolen. Kaunitz''s white rose left on Emma''s dressing table. Catastrophic intelligence loss. Escalated cult alert level.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Beethoven''s Seventh at the Reception', 'event',
 'Performance of Beethoven''s Seventh Symphony (second movement) at the Hofburg Imperial Reception on August 6. The investigators heard the harmonic structure of the Engine within the music.',
 'Triggered SAN checks. Varrio suffered temporary insanity. Pivotal horror moment revealing the Engine''s influence permeates Vienna''s music.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'horror']),

(36, 'Kali Puja 1814', 'event',
 'The night of October 25, 1814 when Kali Puja overlaps with Diwali on the new moon. Provides sensory and spiritual cover for the Calcutta cult''s Cycle of Dissolution ritual.',
 'Ritual timing for the Calcutta cell. First clear night after monsoon — stars become visible for cosmic alignment.',
 'AUTHORITATIVE', ARRAY['chapter-4']),

(36, 'Grand Masquerade at Palais Lobkowitz', 'event',
 'Masquerade ball on August 8 at Palais Lobkowitz (Eroica Hall). All cult leadership attends. Major social climax with masks providing both opportunity and danger.',
 'Planned for Session 5. Potential Trauttmansdorff extraction. Adler deepens relationship with Anna. Key set-piece.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'upcoming']);

-- ============================================================================
-- Rituals
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'The Grand Canticle', 'ritual',
 'The Aeternum Choir''s cosmic ritual requiring 8 simultaneous performances at sacred sites worldwide. 5 of 8 segments must succeed before the summer solstice 1815 for the Final Chorus to tear open the dimensional veil for Yog-Sothoth.',
 'Three segments disrupted (London, Lyon, Venice). Five cells remain active (Vienna, Warsaw, Luxor, Calcutta, Chengdu/Ouro Preto). Vienna MUST succeed per Campaign Design Notes Decision 001. Distinct from "The Grand Canticle Score" (the physical artifact).',
 'AUTHORITATIVE', ARRAY['campaign-wide', 'mythos']),

(36, 'Calcutta Ritual', 'ritual',
 'The Calcutta cell''s Cycle of Dissolution segment. Built around funerary chants and river resonance, timed to Kali Puja on October 25, 1814.',
 'Exploits Hooghly River acoustics and cremation ghat energy. Ritual window aligned with the first clear night after monsoon.',
 'AUTHORITATIVE', ARRAY['chapter-4']);

-- ============================================================================
-- Organizations
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Polizeidirektion', 'organization',
 'Vienna''s official police administration. Handles foreign registration, civil order, and criminal investigation. Exploited by Inspektor Vogel for the Palais Kinsky room search.',
 'Vogel uses its legitimate authority as cover for cult operations. Distinct from the Geheimpolizei (secret police).',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'East India Company', 'organization',
 'The British trading company that rules Calcutta in 1814. Controls military, political, and social life.',
 'Runs regular sailings that investigators will use to travel from Trieste to India. Source of passage and social connections in Chapter 4.',
 'AUTHORITATIVE', ARRAY['chapter-4']),

(36, 'Confraternita dell''Acqua Ferma', 'cult',
 'The Brotherhood of Still Water — Venetian Aeternum Choir cell that attempted to awaken the drowned dead through the Stillwater Choir ritual beneath San Trionfo.',
 'Destroyed by Varrio Harrowmont. Led by Contessa Erminia Foscarini. The Venice ritual (Segment III) was disrupted but the city is described as "no longer truly asleep."',
 'AUTHORITATIVE', ARRAY['chapter-2', 'destroyed']);

-- ============================================================================
-- Chapter-Entity Linkages
-- ============================================================================
-- Link entities to the chapters where they appear.
-- Chapter IDs: 2=London, 3=Lyon, 4=Venice, 5=Vienna, 6=Calcutta

-- London chapter entities
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 2, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Augustus Bolt', 'Jacob', 'Dr. Julius Hume', 'Lady Aurelia Danforth',
  'Kensington Stables', 'Stonehenge Ritual', 'Herzfeld''s Letter to Hume'
);

-- Lyon chapter entities
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 3, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Augustus Bolt', 'Jacob', 'Adrien, Viscount de Montferrand',
  'Comte Emeric de Puyrault', 'Sabine Lambert', 'Claudine Morel',
  'Victoire Lenoir', 'Lucien Goupil', 'Magistrate Beraud',
  'Vicomte Henri Lavigne', 'Madame Leontine des Jardins',
  'Chakota', 'Ciimba', 'Formless Spawn', 'Aeternox', 'Tsathoggua',
  'Paris', 'Le Coteau des Ombres', 'Silkweavers Guild',
  'Orphans Hospital', 'Temple of Aeternox', 'Maison du Corbeau',
  'Fourviere Hill', 'The Orphic Lyre', 'Carreau''s Ledger of Children',
  'Orphans Hospital Raid', 'July 1st Masked Soiree'
);

-- Venice chapter entities
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 4, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Contessa Erminia Foscarini', 'San Trionfo', 'Awakened Drowned',
  'Confraternita dell''Acqua Ferma'
);

-- Vienna chapter entities
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 5, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Adrien, Viscount de Montferrand',
  'Lady Ashworth', 'Freddie Cavendish', 'Klaus Bauer', 'Werner Bauer',
  'Widow Katz', 'Mr. Hartley', 'Caroline Hartley', 'Lydia Hartley',
  'Major Konstantin Thurner', 'Sergeant Franz Huber', 'Greta Adler',
  'Dr. Gottfried Reiner', 'Herr Gustav Metzger', 'Fraulein Liesel Hartmann',
  'Frau Ingrid Sperl', 'Henri Valmont', 'Herr Braun',
  'Dmitri Sokolov', 'Alexei Petrov', 'Ivan Markov', 'Sergei Orlov',
  'Nightgaunt',
  'Palais Kinsky', 'Cafe Frauenhuber', 'Josefstadt District',
  'Vienna Conservatory', 'Eroica Hall',
  'Liber Ivonis', 'Cultes des Goules', 'De Vermis Mysteriis',
  'Revelations of Glaaki',
  'Marina''s Notebook', 'Dorothea''s Intelligence Dossier', 'Volkonsky''s Notebook',
  'Palais Kinsky Room Search', 'Beethoven''s Seventh at the Reception',
  'Grand Masquerade at Palais Lobkowitz'
);

-- Calcutta chapter entities
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 6, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Kali', 'Kalighat Temple', 'Hooghly River', 'Fort William',
  'Kali Puja 1814', 'Calcutta Ritual', 'East India Company'
);

-- ============================================================================
-- Relationships
-- ============================================================================
-- Relationship type IDs (from relationship_types table):
-- 775=bound_to, 787=created, 793=employs, 795=enemy_of, 800=formerly_played,
-- 801=founded, 803=friend_of, 807=headquartered_at, 815=leads, 816=located_at,
-- 818=member_of, 820=murdered, 824=opposes, 826=owns, 827=parent_of,
-- 828=part_of, 829=participated_in, 836=protects, 839=reports_to,
-- 841=rival_of, 844=serves, 845=sibling_of, 848=studies, 858=wields,
-- 860=works_for, 861=worships, 862=wounded

-- PC relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Order of St. Aelfric recruit'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Augustus Bolt'
  AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 829, 'Killed during the Orphans Hospital raid'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Augustus Bolt'
  AND t.campaign_id = 36 AND t.name = 'Orphans Hospital Raid';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 829, 'Killed during the Orphans Hospital raid'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Jacob'
  AND t.campaign_id = 36 AND t.name = 'Orphans Hospital Raid';

-- Adrien relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 803, 'Romantic rapport developing'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Adrien, Viscount de Montferrand'
  AND t.campaign_id = 36 AND t.name = 'Caroline Hartley';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 803, 'Introduced Adrien to the investigators in Lyon'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Comte Emeric de Puyrault'
  AND t.campaign_id = 36 AND t.name = 'Adrien, Viscount de Montferrand';

-- London NPC relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'High cantor and ritualist'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Julius Hume'
  AND t.campaign_id = 36 AND t.name = 'Orphean Society';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Second-in-command and soprano'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lady Aurelia Danforth'
  AND t.campaign_id = 36 AND t.name = 'Orphean Society';

-- Lyon cult relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Priestess of the Lyre'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Sabine Lambert'
  AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Watcher and infiltrator'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Claudine Morel'
  AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 807, 'Cult headquarters with underground chambers'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Societe Harmonique de l''Aube'
  AND t.campaign_id = 36 AND t.name = 'Silkweavers Guild';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Kept in basement behind guarded screen'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Chakota'
  AND t.campaign_id = 36 AND t.name = 'Silkweavers Guild';

-- Venice relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'Conductor of the Stillwater Choir'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Contessa Erminia Foscarini'
  AND t.campaign_id = 36 AND t.name = 'Confraternita dell''Acqua Ferma';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 858, 'Used to control the Chakota'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Contessa Erminia Foscarini'
  AND t.campaign_id = 36 AND t.name = 'The Orphic Lyre';

-- Vienna NPC relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 836, 'Social patron and chaperone'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Lady Ashworth'
  AND t.campaign_id = 36 AND t.name = 'Emma Wentworth';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Brotherhood enforcer'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Klaus Bauer'
  AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Brotherhood enforcer'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Werner Bauer'
  AND t.campaign_id = 36 AND t.name = 'Brotherhood of the Open Measure';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 845, 'Twin brothers'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Klaus Bauer'
  AND t.campaign_id = 36 AND t.name = 'Werner Bauer';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Brenner''s lodging — yellow door'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Wilhelm Brenner'
  AND t.campaign_id = 36 AND t.name = 'Widow Katz';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 827, 'Father'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mr. Hartley'
  AND t.campaign_id = 36 AND t.name = 'Caroline Hartley';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 827, 'Father'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mr. Hartley'
  AND t.campaign_id = 36 AND t.name = 'Lydia Hartley';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Safe house contact and logistics'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Major Konstantin Thurner'
  AND t.campaign_id = 36 AND t.name = 'Order of St. Aelfric';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 839, 'Bribed mercenary informant'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Sergeant Franz Huber'
  AND t.campaign_id = 36 AND t.name = 'Baron Otto von Kaunitz';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 839, 'Reports all investigator movements'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Herr Braun'
  AND t.campaign_id = 36 AND t.name = 'Inspektor Heinrich Vogel';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 845, 'Sister — integrated into Engine soprano register'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Greta Adler'
  AND t.campaign_id = 36 AND t.name = 'Kapellmeister Friedrich Adler';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 775, 'Voice box integrated into soprano register'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Greta Adler'
  AND t.campaign_id = 36 AND t.name = 'The Harmonic Engine';

-- Brotherhood outer circle relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Produces neurotoxins and preservation fluids'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Dr. Gottfried Reiner'
  AND t.campaign_id = 36 AND t.name = 'Professor Albin Herzfeld';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Builds mechanical Engine components'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Herr Gustav Metzger'
  AND t.campaign_id = 36 AND t.name = 'Professor Albin Herzfeld';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Manages victim sedation'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Fraulein Liesel Hartmann'
  AND t.campaign_id = 36 AND t.name = 'Professor Albin Herzfeld';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Maintains the sealed anatomical theatre'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Frau Ingrid Sperl'
  AND t.campaign_id = 36 AND t.name = 'Kapellmeister Friedrich Adler';

-- Der Kantor coordination
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Central European coordinator'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Der Kantor'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Based in Munich'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Der Kantor'
  AND t.campaign_id = 36 AND t.name = 'Munich';

-- Stolen items
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 829, 'Stolen during the room search'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Marina''s Notebook'
  AND t.campaign_id = 36 AND t.name = 'Palais Kinsky Room Search';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 829, 'Stolen during the room search'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Liber Ivonis'
  AND t.campaign_id = 36 AND t.name = 'Palais Kinsky Room Search';

-- Location relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'District within Vienna'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Josefstadt District'
  AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Palais Lobkowitz'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Eroica Hall'
  AND t.campaign_id = 36 AND t.name = 'Palais Lobkowitz';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Lyon'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Fourviere Hill'
  AND t.campaign_id = 36 AND t.name = 'Lyon';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Beneath Fourviere Hill'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Temple of Aeternox'
  AND t.campaign_id = 36 AND t.name = 'Fourviere Hill';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Calcutta'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Kalighat Temple'
  AND t.campaign_id = 36 AND t.name = 'Calcutta';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'River through Calcutta'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Hooghly River'
  AND t.campaign_id = 36 AND t.name = 'Calcutta';

-- Ritual relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 861, 'Ultimate target of the Grand Canticle'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'The Grand Canticle'
  AND t.campaign_id = 36 AND t.name = 'Yog-Sothoth';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Segment within the Grand Canticle'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'The August 15 Ritual'
  AND t.campaign_id = 36 AND t.name = 'The Grand Canticle';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Segment within the Grand Canticle'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Calcutta Ritual'
  AND t.campaign_id = 36 AND t.name = 'The Grand Canticle';

COMMIT;
