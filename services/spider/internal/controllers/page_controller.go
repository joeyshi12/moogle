package controllers

import(
    "log"
    "github.com/redis/go-redis/v9"

    "github.com/IonelPopJara/search-engine/services/spider/internal/pages"
    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
    "github.com/IonelPopJara/search-engine/services/spider/internal/crawler"
    "github.com/IonelPopJara/search-engine/services/spider/internal/database"
)

type PageController struct {
    db *database.Database
}

func NewPageController(db *database.Database) *PageController {
    return &PageController{
        db: db,
    }
}

func (pgc *PageController) GetAllPages() map[string]*pages.Page {
    log.Printf("Fetching data from Redis...\n")
    redisPages := make(map[string]*pages.Page)

    keys, err := pgc.db.Client.Keys(pgc.db.Context, utils.PagePrefix + ":*").Result()
    if err != nil {
        log.Printf("Error fetching data from Redis: %v\n", err)
        return nil
    }

    // Process the redis data using a pipeline
    pipeline := pgc.db.Client.Pipeline()
    cmds := make([]*redis.MapStringStringCmd, len(keys))

    for i, key := range keys {
        cmds[i] = pipeline.HGetAll(pgc.db.Context, key)
    }

    _, err = pipeline.Exec(pgc.db.Context)
    if err != nil {
        log.Printf("Error fetching data from Redis pipeline: %v" , err)
        return nil
    }

    for _, cmd := range cmds {
        data, err := cmd.Result()
        if err != nil {
            log.Printf("Error fetching pipeline result: %v", err)
            return nil
        }

        page, err := pages.DehashPage(data)
        if err != nil {
            log.Printf("Error dehashing page from Redis: %v", err)
            return nil
        }

        redisPages[page.NormalizedURL] = page
    }

    return redisPages
}

func (pgc *PageController) SavePages(crawcfg *crawler.CrawlerConfig) {
    data := crawcfg.Pages
    log.Printf("Writing %d entries to the db...\n", len(data))

    // Process the redis entries using a pipeline
    pipeline := pgc.db.Client.Pipeline()

    for _, page := range data {
        // Check if the page has not been crawled
        // if _, exists := crawcfg.CachedPages[page.NormalizedURL]; !exists {
        // } else {
        //     log.Printf("Skipping %v. Already crawled\n", page.NormalizedURL)
        // }
        // Hash page
        pageHash, err := pages.HashPage(page)
        if err != nil {
            log.Printf("Error hashing page %s: %v", page.NormalizedURL, err)
            continue
        }

        // Append command to pipeline
        pipeline.HSet(pgc.db.Context, utils.PagePrefix + ":"+page.NormalizedURL, pageHash)

        // Push to the indexer queue
        // NOTE: For some weird reason "indexer_queue" does not work, but any other name does :/
        // res, err := pgc.db.Client.LPush(pgc.db.Context, utils.IndexerQueueKey, utils.PagePrefix + ":" +page.NormalizedURL).Result()
        pgc.db.Client.LPush(pgc.db.Context, utils.IndexerQueueKey, utils.PagePrefix + ":" +page.NormalizedURL).Result()

    }

    // Execute the pipeline
    _, err := pipeline.Exec(pgc.db.Context)
    if err != nil {
        log.Printf("Error executing pipeline: %v", err)
    } else {
        log.Printf("Successfully written %d entries to the db!", len(data))
    }
}
