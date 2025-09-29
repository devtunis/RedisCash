version 0.0
app.get("/v1", async (req, res) => {
  try {
    const exists = await client.exists("allpost_hash");

    if (exists) {
   
      const hash = await client.hGetAll("allpost_hash");
     
      const result = Object.values(hash).map(v => JSON.parse(v));
      console.log("🚀 send from Redis hash (fast)");
      return res.status(200).json(result);
    } else {
  
      const data = await PieceTable.find({});

   
      data.forEach((item, index) => {
        client.hSet("allpost_hash", index, JSON.stringify(item));
      });

      console.log("💾 send from MongoDB and saved to Redis hash");
      return res.status(200).json(data);
    }
  } catch (error) {
    return res.status(500).json({ message: error.message });
  }
});
version 0.1
 app.get("/v1", async (req, res) => {
       try{ 
        

           const EXPLORE = await client.get("allpostredis") 
           if(EXPLORE==null){
             const data = await PieceTable.find({})
             await client.set("allpostredis",JSON.stringify(data))
             res.status(200).json(data)
             console.log("send from database ")
           }else{
            console.log("send from cash ")
            const parseExplore = JSON.parse(EXPLORE)
            res.status(200).json(parseExplore)

           }
        
  

       }catch(erorr){
          res.status(404).json({message : erorr.message})
       }
});
