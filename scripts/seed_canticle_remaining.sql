-- ============================================================================
-- Canticle Of The End — Remaining Entities
-- ============================================================================
-- Campaign ID: 36
-- This adds all remaining named entities from the agent extractions that
-- were not included in the initial seed or the first supplement.
--
-- Run with:
--   docker exec -i imagineer-postgres psql -U imagineer -d imagineer \
--     < scripts/seed_canticle_remaining.sql
-- ============================================================================

BEGIN;

-- ============================================================================
-- NPCs — London (Chapter 1) minor characters
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Imogen Bellamy', 'npc',
 'Order of St. Aelfric member captured by the Orphean Society and used as a sacrifice at the Stonehenge ritual.',
 'Deceased. Bound and sacrificed during the Stonehenge ritual.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'order-of-st-aelfric']),

(36, 'Nathaniel Creed', 'npc',
 'Order of St. Aelfric member captured by the Orphean Society and used as a sacrifice at the Stonehenge ritual.',
 'Deceased. Bound and sacrificed during the Stonehenge ritual.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'order-of-st-aelfric']),

(36, 'Clara Fen', 'npc',
 'Student at Hume''s academy who hid a Canticle fragment near her bunk.',
 'Unknown status — likely dead or missing after the cell collapsed.',
 'AUTHORITATIVE', ARRAY['chapter-1', 'orphean-society']),

(36, 'Viola Atwood', 'npc',
 'Student found dead after attending an Orphean Society-sponsored vocal retreat. Age 22, of Wapping.',
 'Deceased. Victim of Orphean Society experiments. Her death is evidence of the cult''s London activities.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-1', 'orphean-society']);

-- ============================================================================
-- NPCs — Lyon (Chapter 2) allies and neutrals
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Etienne de Puyrault', 'npc',
 'Emeric de Puyrault''s younger brother. Missing for 6 months, possibly victim of bandits or the Societe Harmonique.',
 'Presumed dead. His signet ring was found by investigators.',
 'AUTHORITATIVE', ARRAY['missing', 'chapter-2']),

(36, 'Monsieur Albert Lambert', 'npc',
 'Antiquarian, melancholic composer and pianist. Attended the Paris Ball at Hotel de Brissac. Knows about the Orphic Lyre from personal experience.',
 'Alive. Informant. Distinct from Sabine Lambert (Priestess of the Lyre) and Annette Lambert (archivist).',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-2']),

(36, 'Madame Sylvie de Noyelles', 'npc',
 'Married flirt at the Paris Ball. Part of Lyon aristocracy.',
 'Alive. Gossip and social colour.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Madame Claudine Bertier', 'npc',
 'Innkeeper of the Auberge du Griffon Bleu in central Lyon. Broad-hipped woman in her fifties.',
 'Alive. Wife of Alphonse Bertier.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Alphonse Bertier', 'npc',
 'Former coachman, husband of Claudine. Mostly deaf. Tends the books at the Griffon Bleu inn.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Mathilde Bertier', 'npc',
 'Daughter of the Griffon Bleu innkeepers. Serves meals and is the source of all local gossip in Lyon.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Madame Delphine Maret', 'npc',
 'Owner of Le Jardin des Murmures cafe. Soft-spoken widow with ink-stained fingers.',
 'Alive. May know more than she admits.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Jules', 'npc',
 'Steward to Comte de Puyrault at the Le Coteau des Ombres estate. Mid-50s, cautious and gruff.',
 'Alive. Loyal servant of the Puyrault family.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']),

(36, 'Marquis Etienne de Valombre', 'npc',
 'Senior noble of ancient Lugdunum lineage. Father of rescued child Aimee.',
 'Alive. Powerful potential ally against Savarin after the hospital raid.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Madame Therese Bouchard', 'npc',
 'Master of the Silkweavers Guild — the first woman to hold the rank. Mother of rescued son Luc.',
 'Alive. Controls knowledge of underground spaces used by cult. Potential ally.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Councillor Pierre-Henri Montrelais', 'npc',
 'Member of the Municipal Council of Lyon. Father of rescued adopted daughter Colette.',
 'Alive. Unwittingly approved cult excavation funding. Potential ally after rescue.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Baroness Genevieve de Roquemerle', 'npc',
 'Influential Lyon salonniere. Mother of rescued son Julien.',
 'Alive. Controls reputation and elite invitations in Lyon. Potential ally.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Doctor Alexandre Fournier', 'npc',
 'Professor of Medicine at Universite de Lyon. Uncle of rescued niece Sandrine.',
 'Alive. Had provided scientific cover for Doctor Carreau. Potential ally.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Captain Armand Delacourt', 'npc',
 'Retired cavalry officer and commander of a local veterans association. Grandfather of rescued grandson Henri.',
 'Alive. His veterans were the recruitment pool for Savarin''s blue-sashed guards. Potential ally.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Abbess Marie-Clotilde du Prel', 'npc',
 'Abbess of Sainte-Elise Convent. Guardian of Roman-era Aeternox manuscripts.',
 'Alive. Guards historical records relevant to the Temple of Aeternox.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Lord Rupert Granville', 'npc',
 'British diplomat and envoy in Lyon. Father of rescued 14-year-old son.',
 'Alive. Raises international stakes. Potential ally.',
 'AUTHORITATIVE', ARRAY['ally', 'chapter-2']),

(36, 'Marquis Etienne Bravon', 'npc',
 'Notable NPC encountered during the investigators'' journey toward Lyon.',
 'Alive. Met en route to Lyon.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-2']);

-- ============================================================================
-- NPCs — Lyon (Chapter 2) cult members
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Emile Fouchard', 'npc',
 'Cult agent, musician, composer, and vocal coach. Handler and conspirator paired with Victoire Lenoir at the soiree.',
 'Alive. Distinct from Captain Luc Fouchard (military enforcer / blue-sash leader). GM should confirm these are separate characters.',
 'DRAFT', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Antoine Lefevre', 'npc',
 'Cult bureaucrat operating within Lyon''s civil infrastructure. Handles permits and misdirects officials investigating cult activities.',
 'Alive. Savarin''s bureaucratic shield. True believer.',
 'AUTHORITATIVE', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Comte Henri Vallin', 'npc',
 'Occult financier and aristocratic sponsor of the Societe Harmonique. Provides funding and recruits patrons.',
 'Alive. Attended July 1st soiree. Carries hidden ritual tokens.',
 'AUTHORITATIVE', ARRAY['cult-member', 'chapter-2', 'societe-harmonique']),

(36, 'Annette Lambert', 'npc',
 'Music librarian and archivist. Cult sympathiser, possibly under subtle supernatural influence.',
 'Alive. Attended soiree. Distinct from Sabine Lambert (inner circle Priestess of the Lyre).',
 'AUTHORITATIVE', ARRAY['cult-sympathiser', 'chapter-2']),

(36, 'Abbe Ferrant', 'npc',
 'High-ranking cleric in Lyon. Rival of Abbe Duplessis. Fanatic believer in Savarin.',
 'Alive. Savarin''s religious/social shield. Enemy of Duplessis.',
 'AUTHORITATIVE', ARRAY['cult-member', 'chapter-2']),

(36, 'Prefet Etienne Merle', 'npc',
 'Prefet du Rhone. Bribed and compromised by Savarin.',
 'Alive. Controls the Prefecture. Savarin''s man in government.',
 'AUTHORITATIVE', ARRAY['antagonist', 'chapter-2', 'corrupted']),

(36, 'Benoit Chasseloup', 'npc',
 'Silkweaver and tunnel mapmaker. Provided maps of the tunnel system beneath Croix-Rousse to investigators.',
 'Unknown status — deceased or missing. His maps may be misleading.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Mathieu Leclerc', 'npc',
 'Doctor Carreau''s younger apprentice. Handles scheduling and public correspondence.',
 'Alive — status uncertain after Carreau''s death in the hospital raid.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Sister Marguerite', 'npc',
 'Ghost-like figure. Former musical tutor at a children''s school who vanished in 1812. Appears before supernatural events.',
 'Dead/undead apparition. Appears in silkweaver tunnels. Optional encounter.',
 'AUTHORITATIVE', ARRAY['undead', 'chapter-2']);

-- ============================================================================
-- NPCs — Vienna (Chapter 3) minor characters
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Pemberton', 'npc',
 'Freddie Cavendish''s competent and long-suffering manservant. Arranges logistics for the party''s masquerade access.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Mrs. Hartley', 'npc',
 'Wife of Mr. Hartley. Pleased by Adrien''s continental hand-kissing customs.',
 'Alive. Mother of Caroline and Lydia.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Princess Bagration', 'npc',
 'Notorious Russian society figure who wears red to cause a scene at Viennese events. Knows everyone''s scandals.',
 'Alive. Part of the Russian delegation.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Prince Razumovsky', 'npc',
 'Wealthy Russian patron of music. Part of the Russian delegation at the opera.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Prince Nikolaus Esterhazy', 'npc',
 'Head of the Esterhazy family. Described as a fool by Lady Ashworth but immensely wealthy — the richest family after the Habsburgs.',
 'Alive. Patron of musicians.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Liesl', 'npc',
 'Maid at Palais Kinsky. Fetched brothers'' clothes for Charlotte and Georgiana''s disguise. Bribed by police to report on investigators'' movements.',
 'Alive. Her cousin''s friend disappeared after auditioning for Herzfeld.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Little Hansi', 'npc',
 'Street urchin who can be recruited as messenger, spy, and guide through Vienna''s streets.',
 'Alive. Recurring minor contact for the party.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Matthias the Blind Fiddler', 'npc',
 'Former conservatory student who lost his sight after attending one of Herzfeld''s demonstrations. Plays violin on the streets and speaks in riddles.',
 'Alive. Cryptic informant. Knows fragments of truth about the Engine.',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-3']),

(36, 'Father Matthias', 'npc',
 'Priest at St. Stephen''s Cathedral troubled by confessions he has been hearing.',
 'Alive. Cannot break the seal of confession but can hint at terrible knowledge being shared.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Dr. Goldstein', 'npc',
 'Jewish physician and community leader in Leopoldstadt. Knows Dr. Brenner and has treated cult victims who escaped.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-3']),

(36, 'Signora Benedetti', 'npc',
 'Theatrical costumier who knows everyone who attends masked events. Can identify people by their costumes.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['informant', 'chapter-3']),

(36, 'Herr Dietrich', 'npc',
 'Reliable and discreet cabman who drove the party to Leopoldstadt. Knows they went to the Crooked Chimney area.',
 'Alive. Available as recurring transport.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Signor Morosi', 'npc',
 'Gracious older gentleman at the Italian-Sardinian delegation. Bonded with Varrio over mustache wax and provided an Imperial Reception invitation.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Mr. Pembroke', 'npc',
 'Thin, precise secretary to the British consul in Vienna. Initially refused the investigators entry before relenting.',
 'Alive.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Fraulein Maria Holzer', 'npc',
 'Young soprano, age 22. Early Harmonic Engine victim who died during "integration."',
 'Deceased. Family paid for silence. Official cause listed as consumption. Distinct from Frau Margarethe Holzer.',
 'AUTHORITATIVE', ARRAY['deceased', 'chapter-3', 'engine-victim']),

(36, 'Baroness von Helstein', 'npc',
 'Minor noblewoman whose runaway carriage horse can be stopped by the party — a potential source of introductions and gossip if rescued.',
 'Alive. Sandbox encounter NPC.',
 'AUTHORITATIVE', ARRAY['civilian', 'chapter-3']),

(36, 'Talleyrand', 'npc',
 'Legendary French diplomat leading the French delegation in Vienna. Dorothea de Courlande''s uncle.',
 'Alive. Controls French intelligence network.',
 'AUTHORITATIVE', ARRAY['historical', 'chapter-3']),

(36, 'Dr. Septimus Harrowmont', 'npc',
 'Former British Army field surgeon and Order of St. Aelfric consulting surgeon. Dispatched to Venice where he disrupted the Confraternita del Bel Canto from within.',
 'Alive. NPC precursor to the PC Varrio Harrowmont — may be the same character or a relative depending on GM preference.',
 'DRAFT', ARRAY['chapter-2', 'order-of-st-aelfric']);

-- ============================================================================
-- Locations — London
-- ============================================================================
-- (Kensington Stables already added in supplement)

-- ============================================================================
-- Locations — Lyon (Chapter 2)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Auberge du Griffon Bleu', 'location',
 'Coaching inn in central Lyon near Place des Terreaux. Investigators'' first lodging in Lyon.',
 'Run by the Bertier family.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Croix-Rousse District', 'location',
 'Hilly, working-class neighbourhood of silk weavers (canuts) in Lyon. Location of the tunnel entrance to the cult''s underground network.',
 'Contains the Silkweavers Guild and tunnel access via Rue des Tisserands.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Le Jardin des Murmures', 'location',
 'Discreet cafe near Place Bellecour in Lyon. Investigators'' study location for Mythos tomes and spells.',
 'Run by Madame Delphine Maret.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Maison du Choeur Sacre', 'location',
 'Private salon in Lyon''s Presqu''ile district south of Place Bellecour. Hosted by Madame des Jardins for Societe Harmonique events.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Montferand''s Hunting Lodge', 'location',
 'Converted Roman watchtower on northern slopes above Lyon overlooking the Rhone valley. Contains Adrien''s occult library.',
 'Owned by Adrien, Viscount de Montferrand.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Hotel-Dieu de Lyon', 'location',
 'One of Lyon''s oldest hospitals. Doctor Carreau''s public workplace.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Collegiale Saint-Just', 'location',
 'Church in Lyon where Abbe Duplessis serves as librarian-priest and Order of St. Aelfric operative.',
 'Duplessis''s cover. Referenced in the code phrase "the librarian of Saint-Just."',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Sainte-Elise Convent', 'location',
 'Convent in Lyon guarding historical records including Roman-era manuscripts describing the worship of Aeternox.',
 'Abbess Marie-Clotilde du Prel is guardian.',
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Place Bellecour', 'location',
 'Major square in Lyon''s Presqu''ile district.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Place des Terreaux', 'location',
 'Square in Lyon near the Hotel de Ville. Near the Auberge du Griffon Bleu.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']),

(36, 'Presqu''ile District', 'location',
 'Central peninsula area of Lyon between the Saone and Rhone rivers. Location of high society venues.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-2']);

-- ============================================================================
-- Locations — Vienna (Chapter 3)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Palais Thun-Hohenstein', 'location',
 'Residence of Countess Maria von Thun near Herrengasse. Site of her famous salons where Sternberg and Kaunitz encounters occur.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Am Hof Square', 'location',
 'Public square outside Palais Kinsky with a fountain. Watched by Geheimpolizei brown-coat agents.',
 'Herr Braun''s surveillance post.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Pension Muller', 'location',
 'Modest pension on Josefstadter Strasse 12 where Madame Celestine Delacroix stays in Vienna.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Palais Kaunitz (French Embassy)', 'location',
 'Building housing the French delegation in Vienna. Site of Dorothea de Courlande''s salon.',
 'Different from Baron von Kaunitz''s personal residence.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Great Redoutensaal', 'location',
 'Grand ballroom in the Hofburg Palace. Site of the Imperial Reception.',
 'Where the Harcourt reunion and Beethoven SAN-check moment occurred.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Pension Gruber', 'location',
 'Anna Lindqvist''s lodging in the Alsergrund district near the University of Vienna.',
 'Known to Adler for soprano acquisition planning.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Palais Trauttmansdorff', 'location',
 'Count Trauttmansdorff''s residence in the Innere Stadt showing signs of decay. His study contains travel documents and evidence of planned flight.',
 'Cult meeting location. Potential extraction point for Trauttmansdorff.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Lange Gasse 14', 'location',
 'Major Thurner''s apartment in Josefstadt district. The Order of St. Aelfric''s safe house in Vienna.',
 'Weapons cache, medical supplies, secure fallback position.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'order-of-st-aelfric']),

(36, 'Belvedere Gallery', 'location',
 'Art gallery in Vienna where Charlotte and Adrien spent a morning. Reading room contains periodicals.',
 'Where Charlotte found the Wiener Medicinische Zeitung clipping about Herzfeld''s grant.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Palais Esterhazy', 'location',
 'Residence of the Esterhazy family. Venue for Schubert Lieder recitals.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Innere Stadt', 'location',
 'Vienna''s inner city district. The fashionable centre where most society events and key buildings are located.',
 'Herzfeld''s private apartments are here.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Alsergrund', 'location',
 'District near the University of Vienna where Anna Lindqvist lodges at Pension Gruber.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Theater an der Wien', 'location',
 'Theatre near theatrical costumiers in Vienna. Hosts Italian opera and comic opera.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Augarten Pavilion', 'location',
 'Public concert venue in Vienna. Site of military band concerts open to all.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']);

-- ============================================================================
-- Artifacts and Items
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Carreau''s Harmonic Tuning Fork', 'artifact',
 'Dark-metal tuning fork that produces unsettling after-tones. Used by Doctor Carreau to test resonance reactions in children.',
 'Status uncertain after Carreau''s death.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']),

(36, 'Chasseloup''s Tunnel Map', 'artifact',
 'Map of the silkweaver tunnels beneath Croix-Rousse provided by Benoit Chasseloup. Shows route to the Temple of Aeternox excavation.',
 'In investigators'' possession. Maps may be misleading or incomplete.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-2']),

(36, 'Brotherhood Wax Seal', 'artifact',
 'Wax seal bearing a tuning fork surrounded by a spiral. Used on Brotherhood of the Open Measure cult correspondence.',
 'Observed on Kaunitz''s correspondence at Cafe Frauenhuber.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-3']),

(36, 'Harcourt''s Coded Letter', 'artifact',
 'Intercepted and partially decoded correspondence showing the eight Aeternum Choir cells and the five-of-eight requirement for the Grand Canticle.',
 'In Lord Harcourt''s possession. Reveals the full scope of the global plan.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-3']),

(36, 'False Arrest Warrants', 'artifact',
 'Prepared warrants charging the investigators with espionage, theft of state documents, and conspiracy against the Crown.',
 'Held by Kaunitz. Prepared by Vogel. Can be deployed at any moment as leverage against the party.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'threat']);

-- ============================================================================
-- Documents
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Danforth''s Private Notes', 'document',
 'Notes revealing Lady Danforth''s resentment of Hume and ambitions to control the Canticle independently.',
 'Found at Orphean Society. Chapter 1 evidence.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-1']),

(36, 'Savarin''s Letter to Hume', 'document',
 'Letter from Mathilde Savarin found in Hume''s office. Discusses Canticle segment composition and coordination between Lyon and London cells.',
 'Chapter 1 evidence linking the two cells.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-1']),

(36, 'Wiener Medicinische Zeitung Clipping', 'document',
 'German newspaper snippet confirming Herzfeld received an imperial grant for acoustic research including renovation of anatomical facilities.',
 'In investigators'' possession. Discovered by Charlotte at the Belvedere Gallery reading room.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-3']),

(36, 'Carreau''s Encrypted Letters from Savarin', 'document',
 'Two sealed letters with instructions on which children to preserve for the choir and schedule for transfer to the rehearsal site.',
 'Chapter 2 evidence linking Carreau''s hospital to Savarin''s ritual plans.',
 'AUTHORITATIVE', ARRAY['evidence', 'chapter-2']),

(36, 'Threnody of Nephren-Ka', 'document',
 'Ancient Mythos text fragment describing "the Black Choir beneath the Earth." Referenced by Georgiana during the soiree.',
 'Fragment only. Chapter 2.',
 'AUTHORITATIVE', ARRAY['mythos', 'chapter-2']);

-- ============================================================================
-- Events — Vienna calendar and session events
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Don Giovanni at the Burgtheater', 'event',
 'Performance of Mozart''s Don Giovanni on August 4. The investigators'' first social event in Vienna.',
 'Past (Session 2). Where the party first sighted Anna Lindqvist with Adler. Kaunitz first approached Emma.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'past']),

(36, 'The Harcourt Reunion', 'event',
 'Meeting between the investigators and Lord Harcourt and Lady Honoria at the Imperial Reception on August 6.',
 'Past (Session 4). Order of St. Aelfric reunited. Led to morning debrief at Palais Modena.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'past']),

(36, 'Anna Lindqvist''s Concert', 'event',
 'Vocal concert at the Burgtheater on August 10 featuring Swedish and German art songs. Anna''s career-making performance.',
 'Future. Adler plans to acquire Anna the morning after this concert.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'upcoming']),

(36, 'Anna Lindqvist''s Recital at von Thun''s', 'event',
 'Private recital at Countess von Thun''s residence on August 8 afternoon.',
 'Future (Session 5). Where PCs can hear the Engine''s hunger in Anna''s voice — major horror beat.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'upcoming', 'horror']),

(36, 'Feast of the Assumption', 'event',
 'Catholic holy day on August 15. The date when the Harmonic Engine ritual is to be performed at midnight.',
 'The ritual''s cosmic alignment date. Campaign Design Notes Decision 001: ritual MUST succeed.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Grand Ball at Schonbrunn', 'event',
 'Imperial invitation-only ball at Schonbrunn Palace on August 12.',
 'Vienna events calendar. Cover event while Anna''s integration into the Engine begins.',
 'AUTHORITATIVE', ARRAY['chapter-3', 'upcoming']),

(36, 'Paris Ball at Hotel de Brissac', 'event',
 'Grand ball in Paris where investigators met Lavigne, Lambert, Noyelles, and others before travelling to Lyon.',
 'Chapter 2 opening event.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'past']);

-- ============================================================================
-- Rituals
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Venice Ritual', 'ritual',
 'The Stillwater Choir ritual held beneath San Trionfo that attempted to awaken Venice''s drowned dead. Segment III of the Grand Canticle.',
 'Disrupted by Varrio Harrowmont. Contessa Foscarini killed. Venice described as "no longer truly asleep" — partial success?',
 'AUTHORITATIVE', ARRAY['chapter-2', 'ritual-disrupted']),

(36, 'Bastille Day Ritual', 'ritual',
 'Savarin''s planned final ritual on July 14 at the Temple of Aeternox beneath Fourviere Hill. Intended to open "Aeternox''s Gate of Duration." Segment II of the Grand Canticle.',
 'Disrupted by investigators. Chapter 2 climax.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'ritual-disrupted']);

-- ============================================================================
-- Organizations
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'French Delegation', 'organization',
 'France''s diplomatic presence in Vienna, housed at the Palais Kaunitz (French Embassy). Source of intelligence via Dorothea and La Tour-du-Pin.',
 'Talleyrand leads.',
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Russian Delegation', 'organization',
 'Russia''s diplomatic presence at the Congress of Vienna. Count Nesselrode leads. Volkonsky provides intelligence on missing Russian musicians.',
 NULL,
 'AUTHORITATIVE', ARRAY['chapter-3']),

(36, 'Blue-Sashed Guards', 'organization',
 'Savarin''s private enforcers in Lyon. Ex-Napoleonic veterans wearing pale-blue sashes as a symbol of loyalty.',
 'Recruited from Captain Delacourt''s veterans association. Led by Captain Luc Fouchard.',
 'AUTHORITATIVE', ARRAY['chapter-2', 'societe-harmonique']);

-- ============================================================================
-- Remaining Aeternum Choir cells (as cults)
-- ============================================================================

INSERT INTO entities (campaign_id, name, entity_type, description, gm_notes, source_confidence, tags)
VALUES
(36, 'Warsaw Cell', 'cult',
 'Active Aeternum Choir cell performing the Percussive Invocation segment. Ritual window late September 1814.',
 'Order of St. Aelfric has agents monitoring but no capacity to intervene.',
 'AUTHORITATIVE', ARRAY['aeternum-choir', 'active']),

(36, 'Luxor Cell', 'cult',
 'Active Aeternum Choir cell performing the Descent Motif segment. Operating through archaeological cover. Ritual window late November 1814.',
 'Part of the five remaining active cells.',
 'AUTHORITATIVE', ARRAY['aeternum-choir', 'active']),

(36, 'Chengdu Cell', 'cult',
 'Possible Aeternum Choir cell. Intelligence is thin.',
 'Unconfirmed. Order still mapping the network.',
 'DRAFT', ARRAY['aeternum-choir', 'unconfirmed']),

(36, 'Ouro Preto Cell', 'cult',
 'Possible Aeternum Choir cell. Intelligence is thin.',
 'Unconfirmed. Order still mapping the network.',
 'DRAFT', ARRAY['aeternum-choir', 'unconfirmed']);

-- ============================================================================
-- Chapter-Entity Linkages for new entities
-- ============================================================================

-- London chapter
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 2, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Imogen Bellamy', 'Nathaniel Creed', 'Clara Fen', 'Viola Atwood',
  'Danforth''s Private Notes', 'Savarin''s Letter to Hume'
);

-- Lyon chapter
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 3, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Etienne de Puyrault', 'Monsieur Albert Lambert', 'Madame Sylvie de Noyelles',
  'Madame Claudine Bertier', 'Alphonse Bertier', 'Mathilde Bertier',
  'Madame Delphine Maret', 'Jules', 'Marquis Etienne de Valombre',
  'Madame Therese Bouchard', 'Councillor Pierre-Henri Montrelais',
  'Baroness Genevieve de Roquemerle', 'Doctor Alexandre Fournier',
  'Captain Armand Delacourt', 'Abbess Marie-Clotilde du Prel',
  'Lord Rupert Granville', 'Marquis Etienne Bravon',
  'Emile Fouchard', 'Antoine Lefevre', 'Comte Henri Vallin',
  'Annette Lambert', 'Abbe Ferrant', 'Prefet Etienne Merle',
  'Benoit Chasseloup', 'Mathieu Leclerc', 'Sister Marguerite',
  'Auberge du Griffon Bleu', 'Croix-Rousse District',
  'Le Jardin des Murmures', 'Maison du Choeur Sacre',
  'Montferand''s Hunting Lodge', 'Hotel-Dieu de Lyon',
  'Collegiale Saint-Just', 'Sainte-Elise Convent',
  'Place Bellecour', 'Place des Terreaux', 'Presqu''ile District',
  'Carreau''s Harmonic Tuning Fork', 'Chasseloup''s Tunnel Map',
  'Carreau''s Encrypted Letters from Savarin', 'Threnody of Nephren-Ka',
  'Blue-Sashed Guards', 'Paris Ball at Hotel de Brissac',
  'Venice Ritual', 'Bastille Day Ritual'
);

-- Vienna chapter
INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
SELECT 5, id, 'featured' FROM entities WHERE campaign_id = 36 AND name IN (
  'Pemberton', 'Mrs. Hartley', 'Princess Bagration', 'Prince Razumovsky',
  'Prince Nikolaus Esterhazy', 'Liesl', 'Little Hansi',
  'Matthias the Blind Fiddler', 'Father Matthias', 'Dr. Goldstein',
  'Signora Benedetti', 'Herr Dietrich', 'Signor Morosi', 'Mr. Pembroke',
  'Fraulein Maria Holzer', 'Baroness von Helstein', 'Talleyrand',
  'Palais Thun-Hohenstein', 'Am Hof Square', 'Pension Muller',
  'Palais Kaunitz (French Embassy)', 'Great Redoutensaal',
  'Pension Gruber', 'Palais Trauttmansdorff', 'Lange Gasse 14',
  'Belvedere Gallery', 'Palais Esterhazy', 'Innere Stadt',
  'Alsergrund', 'Theater an der Wien', 'Augarten Pavilion',
  'Brotherhood Wax Seal', 'Harcourt''s Coded Letter', 'False Arrest Warrants',
  'Wiener Medicinische Zeitung Clipping',
  'French Delegation', 'Russian Delegation',
  'Don Giovanni at the Burgtheater', 'The Harcourt Reunion',
  'Anna Lindqvist''s Concert', 'Anna Lindqvist''s Recital at von Thun''s',
  'Feast of the Assumption', 'Grand Ball at Schonbrunn'
);

-- ============================================================================
-- Additional Relationships
-- ============================================================================

-- Lyon family relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 845, 'Brothers — Etienne is missing'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Comte Emeric de Puyrault'
  AND t.campaign_id = 36 AND t.name = 'Etienne de Puyrault';

-- Lyon cult membership
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Cult agent and vocal coach'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Emile Fouchard'
  AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 818, 'Bureaucratic shield'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Antoine Lefevre'
  AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 799, 'Occult financier and patron'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Comte Henri Vallin'
  AND t.campaign_id = 36 AND t.name = 'Societe Harmonique de l''Aube';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 795, 'Religious rival'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Abbe Ferrant'
  AND t.campaign_id = 36 AND t.name = 'Abbe Duplessis';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 784, 'Corrupted prefet in Savarin''s pocket'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mathilde Savarin'
  AND t.campaign_id = 36 AND t.name = 'Prefet Etienne Merle';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 784, 'Corrupted magistrate in Savarin''s pocket'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mathilde Savarin'
  AND t.campaign_id = 36 AND t.name = 'Magistrate Beraud';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 860, 'Apprentice to Doctor Carreau'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mathieu Leclerc'
  AND t.campaign_id = 36 AND t.name = 'Doctor Carreau';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 815, 'Leader of the blue-sashed guards'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Captain Luc Fouchard'
  AND t.campaign_id = 36 AND t.name = 'Blue-Sashed Guards';

-- Lyon location relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Lyon'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Croix-Rousse District'
  AND t.campaign_id = 36 AND t.name = 'Lyon';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Lyon'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Presqu''ile District'
  AND t.campaign_id = 36 AND t.name = 'Lyon';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Duplessis''s cover as librarian-priest'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Abbe Duplessis'
  AND t.campaign_id = 36 AND t.name = 'Collegiale Saint-Just';

-- Vienna location relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Vienna'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Am Hof Square'
  AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Vienna'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Innere Stadt'
  AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Within Vienna'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Alsergrund'
  AND t.campaign_id = 36 AND t.name = 'Vienna';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Countess von Thun''s residence'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Countess Maria von Thun'
  AND t.campaign_id = 36 AND t.name = 'Palais Thun-Hohenstein';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Trauttmansdorff''s residence'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Count Leopold von Trauttmansdorff'
  AND t.campaign_id = 36 AND t.name = 'Palais Trauttmansdorff';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Anna''s lodging near the University'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Anna Lindqvist'
  AND t.campaign_id = 36 AND t.name = 'Pension Gruber';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Delacroix''s lodging'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Madame Celestine Delacroix'
  AND t.campaign_id = 36 AND t.name = 'Pension Muller';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 816, 'Order safe house'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Major Konstantin Thurner'
  AND t.campaign_id = 36 AND t.name = 'Lange Gasse 14';

-- Vienna NPC relationships
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 844, 'Manservant'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Pemberton'
  AND t.campaign_id = 36 AND t.name = 'Freddie Cavendish';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 827, 'Mother of Caroline and Lydia'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mrs. Hartley'
  AND t.campaign_id = 36 AND t.name = 'Caroline Hartley';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 846, 'Married'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Mrs. Hartley'
  AND t.campaign_id = 36 AND t.name = 'Mr. Hartley';

-- Talleyrand / Dorothea relationship
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 827, 'Uncle (controls French intelligence)'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Talleyrand'
  AND t.campaign_id = 36 AND t.name = 'Dorothea de Courlande';

-- Cell relationships to Aeternum Choir
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Active cell — Percussive Invocation segment'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Warsaw Cell'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Active cell — Descent Motif segment'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Luxor Cell'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Possible cell — unconfirmed'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Chengdu Cell'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Possible cell — unconfirmed'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Ouro Preto Cell'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

-- Note: Orphean Society, Societe Harmonique, Confraternita del Bel Canto,
-- and Brotherhood of the Open Measure already have part_of relationships
-- to Aeternum Choir from earlier seeds. Only adding the new Venice offshoot.
INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id, relationship_type_id, description)
SELECT 36, s.id, t.id, 828, 'Disrupted cell — Segment III (Venice offshoot)'
FROM entities s, entities t
WHERE s.campaign_id = 36 AND s.name = 'Confraternita dell''Acqua Ferma'
  AND t.campaign_id = 36 AND t.name = 'Aeternum Choir';

COMMIT;
