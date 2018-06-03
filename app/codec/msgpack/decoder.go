package main

//func (d *Decoder) queryMapKey(q *queryResult) error {
//	n, err := d.DecodeMapLen()
//	if err != nil {
//		return err
//	}
//	if n == -1 {
//		return nil
//	}
//
//	for i := 0; i < n; i++ {
//		k, err := d.bytesNoCopy()
//		if err != nil {
//			return err
//		}
//
//		if string(k) == q.key {
//			if err := d.query(q); err != nil {
//				return err
//			}
//			if q.hasAsterisk {
//				return d.skipNext((n - i - 1) * 2)
//			}
//			return nil
//		}
//
//		if err := d.Skip(); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (d *Decoder) queryArrayIndex(q *queryResult) error {
//	n, err := d.DecodeArrayLen()
//	if err != nil {
//		return err
//	}
//	if n == -1 {
//		return nil
//	}
//
//	if q.key == "*" {
//		q.hasAsterisk = true
//
//		query := q.query
//		for i := 0; i < n; i++ {
//			q.query = query
//			if err := d.query(q); err != nil {
//				return err
//			}
//		}
//
//		q.hasAsterisk = false
//		return nil
//	}
//
//	ind, err := strconv.Atoi(q.key)
//	if err != nil {
//		return err
//	}
//
//	for i := 0; i < n; i++ {
//		if i == ind {
//			if err := d.query(q); err != nil {
//				return err
//			}
//			if q.hasAsterisk {
//				return d.skipNext(n - i - 1)
//			}
//			return nil
//		}
//
//		if err := d.Skip(); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
