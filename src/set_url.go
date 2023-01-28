package set_url


func set(client *redis.Client, key string, val string)(error){
  err := client.Set(key, val, 0).Err()
  if err != nil {
    return err
  }
  return nil
}
