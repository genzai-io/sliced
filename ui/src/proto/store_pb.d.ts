// package: store_pb
// file: proto/store.proto

import * as jspb from "google-protobuf";

export class App extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getDesc(): string;
  setDesc(value: string): void;

  clearVersionsList(): void;
  getVersionsList(): Array<string>;
  setVersionsList(value: Array<string>): void;
  addVersions(value: string, index?: number): string;

  clearTagsList(): void;
  getTagsList(): Array<string>;
  setTagsList(value: Array<string>): void;
  addTags(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): App.AsObject;
  static toObject(includeInstance: boolean, msg: App): App.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: App, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): App;
  static deserializeBinaryFromReader(message: App, reader: jspb.BinaryReader): App;
}

export namespace App {
  export type AsObject = {
    id: number,
    name: string,
    desc: string,
    versionsList: Array<string>,
    tagsList: Array<string>,
  }
}

export class Worker extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getVersion(): string;
  setVersion(value: string): void;

  getAddress(): string;
  setAddress(value: string): void;

  getMemory(): number;
  setMemory(value: number): void;

  getCpus(): number;
  setCpus(value: number): void;

  clearQueuesList(): void;
  getQueuesList(): Array<string>;
  setQueuesList(value: Array<string>): void;
  addQueues(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Worker.AsObject;
  static toObject(includeInstance: boolean, msg: Worker): Worker.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Worker, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Worker;
  static deserializeBinaryFromReader(message: Worker, reader: jspb.BinaryReader): Worker;
}

export namespace Worker {
  export type AsObject = {
    id: string,
    version: string,
    address: string,
    memory: number,
    cpus: number,
    queuesList: Array<string>,
  }
}

export class Node extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getHost(): string;
  setHost(value: string): void;

  getVersion(): string;
  setVersion(value: string): void;

  getInstanceid(): string;
  setInstanceid(value: string): void;

  getRegion(): string;
  setRegion(value: string): void;

  getZone(): string;
  setZone(value: string): void;

  getCores(): number;
  setCores(value: number): void;

  getMemory(): number;
  setMemory(value: number): void;

  getOs(): string;
  setOs(value: string): void;

  getArch(): string;
  setArch(value: string): void;

  getCpuspeed(): number;
  setCpuspeed(value: number): void;

  getBootstrap(): boolean;
  setBootstrap(value: boolean): void;

  getWebhost(): string;
  setWebhost(value: string): void;

  getApihost(): string;
  setApihost(value: string): void;

  getApiloops(): number;
  setApiloops(value: number): void;

  hasMember(): boolean;
  clearMember(): void;
  getMember(): RaftMember | undefined;
  setMember(value?: RaftMember): void;

  getCreated(): number;
  setCreated(value: number): void;

  getInited(): number;
  setInited(value: number): void;

  getChanged(): number;
  setChanged(value: number): void;

  getDropped(): number;
  setDropped(value: number): void;

  getRemoved(): number;
  setRemoved(value: number): void;

  clearDrivesList(): void;
  getDrivesList(): Array<Drive>;
  setDrivesList(value: Array<Drive>): void;
  addDrives(value?: Drive, index?: number): Drive;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Node.AsObject;
  static toObject(includeInstance: boolean, msg: Node): Node.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Node, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Node;
  static deserializeBinaryFromReader(message: Node, reader: jspb.BinaryReader): Node;
}

export namespace Node {
  export type AsObject = {
    id: string,
    host: string,
    version: string,
    instanceid: string,
    region: string,
    zone: string,
    cores: number,
    memory: number,
    os: string,
    arch: string,
    cpuspeed: number,
    bootstrap: boolean,
    webhost: string,
    apihost: string,
    apiloops: number,
    member?: RaftMember.AsObject,
    created: number,
    inited: number,
    changed: number,
    dropped: number,
    removed: number,
    drivesList: Array<Drive.AsObject>,
  }
}

export class RaftMember extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getAddress(): string;
  setAddress(value: string): void;

  getStatus(): RaftStatus;
  setStatus(value: RaftStatus): void;

  getMembership(): Suffrage;
  setMembership(value: Suffrage): void;

  getSuffrage(): Suffrage;
  setSuffrage(value: Suffrage): void;

  getTerm(): number;
  setTerm(value: number): void;

  getApplied(): number;
  setApplied(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RaftMember.AsObject;
  static toObject(includeInstance: boolean, msg: RaftMember): RaftMember.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RaftMember, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RaftMember;
  static deserializeBinaryFromReader(message: RaftMember, reader: jspb.BinaryReader): RaftMember;
}

export namespace RaftMember {
  export type AsObject = {
    id: string,
    address: string,
    status: RaftStatus,
    membership: Suffrage,
    suffrage: Suffrage,
    term: number,
    applied: number,
  }
}

export class NodeGroup extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  clearMembersList(): void;
  getMembersList(): Array<NodeGroup.Member>;
  setMembersList(value: Array<NodeGroup.Member>): void;
  addMembers(value?: NodeGroup.Member, index?: number): NodeGroup.Member;

  hasSlices(): boolean;
  clearSlices(): void;
  getSlices(): Slice | undefined;
  setSlices(value?: Slice): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NodeGroup.AsObject;
  static toObject(includeInstance: boolean, msg: NodeGroup): NodeGroup.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NodeGroup, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NodeGroup;
  static deserializeBinaryFromReader(message: NodeGroup, reader: jspb.BinaryReader): NodeGroup;
}

export namespace NodeGroup {
  export type AsObject = {
    id: number,
    name: string,
    membersList: Array<NodeGroup.Member.AsObject>,
    slices?: Slice.AsObject,
  }

  export class Member extends jspb.Message {
    getNodeid(): string;
    setNodeid(value: string): void;

    getSuffrage(): Suffrage;
    setSuffrage(value: Suffrage): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Member.AsObject;
    static toObject(includeInstance: boolean, msg: Member): Member.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Member, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Member;
    static deserializeBinaryFromReader(message: Member, reader: jspb.BinaryReader): Member;
  }

  export namespace Member {
    export type AsObject = {
      nodeid: string,
      suffrage: Suffrage,
    }
  }
}

export class Database extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getDescription(): string;
  setDescription(value: string): void;

  getCreated(): number;
  setCreated(value: number): void;

  getChanged(): number;
  setChanged(value: number): void;

  getDropped(): number;
  setDropped(value: number): void;

  getRemoved(): number;
  setRemoved(value: number): void;

  clearSlicesList(): void;
  getSlicesList(): Array<Slice>;
  setSlicesList(value: Array<Slice>): void;
  addSlices(value?: Slice, index?: number): Slice;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Database.AsObject;
  static toObject(includeInstance: boolean, msg: Database): Database.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Database, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Database;
  static deserializeBinaryFromReader(message: Database, reader: jspb.BinaryReader): Database;
}

export namespace Database {
  export type AsObject = {
    id: number,
    name: string,
    description: string,
    created: number,
    changed: number,
    dropped: number,
    removed: number,
    slicesList: Array<Slice.AsObject>,
  }
}

export class SliceID extends jspb.Message {
  getDatabaseid(): number;
  setDatabaseid(value: number): void;

  getSliceid(): number;
  setSliceid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SliceID.AsObject;
  static toObject(includeInstance: boolean, msg: SliceID): SliceID.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SliceID, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SliceID;
  static deserializeBinaryFromReader(message: SliceID, reader: jspb.BinaryReader): SliceID;
}

export namespace SliceID {
  export type AsObject = {
    databaseid: number,
    sliceid: number,
  }
}

export class Slice extends jspb.Message {
  hasId(): boolean;
  clearId(): void;
  getId(): SliceID | undefined;
  setId(value?: SliceID): void;

  clearSlotsList(): void;
  getSlotsList(): Array<SlotRange>;
  setSlotsList(value: Array<SlotRange>): void;
  addSlots(value?: SlotRange, index?: number): SlotRange;

  clearNodesList(): void;
  getNodesList(): Array<SliceNode>;
  setNodesList(value: Array<SliceNode>): void;
  addNodes(value?: SliceNode, index?: number): SliceNode;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Slice.AsObject;
  static toObject(includeInstance: boolean, msg: Slice): Slice.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Slice, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Slice;
  static deserializeBinaryFromReader(message: Slice, reader: jspb.BinaryReader): Slice;
}

export namespace Slice {
  export type AsObject = {
    id?: SliceID.AsObject,
    slotsList: Array<SlotRange.AsObject>,
    nodesList: Array<SliceNode.AsObject>,
  }
}

export class SliceNode extends jspb.Message {
  getNodeid(): string;
  setNodeid(value: string): void;

  hasSliceid(): boolean;
  clearSliceid(): void;
  getSliceid(): SliceID | undefined;
  setSliceid(value?: SliceID): void;

  hasMember(): boolean;
  clearMember(): void;
  getMember(): RaftMember | undefined;
  setMember(value?: RaftMember): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SliceNode.AsObject;
  static toObject(includeInstance: boolean, msg: SliceNode): SliceNode.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SliceNode, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SliceNode;
  static deserializeBinaryFromReader(message: SliceNode, reader: jspb.BinaryReader): SliceNode;
}

export namespace SliceNode {
  export type AsObject = {
    nodeid: string,
    sliceid?: SliceID.AsObject,
    member?: RaftMember.AsObject,
  }
}

export class Rebalance extends jspb.Message {
  getTimestamp(): number;
  setTimestamp(value: number): void;

  clearTasksList(): void;
  getTasksList(): Array<Rebalance.Task>;
  setTasksList(value: Array<Rebalance.Task>): void;
  addTasks(value?: Rebalance.Task, index?: number): Rebalance.Task;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Rebalance.AsObject;
  static toObject(includeInstance: boolean, msg: Rebalance): Rebalance.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Rebalance, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Rebalance;
  static deserializeBinaryFromReader(message: Rebalance, reader: jspb.BinaryReader): Rebalance;
}

export namespace Rebalance {
  export type AsObject = {
    timestamp: number,
    tasksList: Array<Rebalance.Task.AsObject>,
  }

  export class Task extends jspb.Message {
    getFrom(): number;
    setFrom(value: number): void;

    getTo(): number;
    setTo(value: number): void;

    getLow(): number;
    setLow(value: number): void;

    getCount(): number;
    setCount(value: number): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Task.AsObject;
    static toObject(includeInstance: boolean, msg: Task): Task.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Task, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Task;
    static deserializeBinaryFromReader(message: Task, reader: jspb.BinaryReader): Task;
  }

  export namespace Task {
    export type AsObject = {
      from: number,
      to: number,
      low: number,
      count: number,
    }
  }
}

export class Ring extends jspb.Message {
  clearRangesList(): void;
  getRangesList(): Array<SlotRange>;
  setRangesList(value: Array<SlotRange>): void;
  addRanges(value?: SlotRange, index?: number): SlotRange;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Ring.AsObject;
  static toObject(includeInstance: boolean, msg: Ring): Ring.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Ring, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Ring;
  static deserializeBinaryFromReader(message: Ring, reader: jspb.BinaryReader): Ring;
}

export namespace Ring {
  export type AsObject = {
    rangesList: Array<SlotRange.AsObject>,
  }
}

export class SlotRange extends jspb.Message {
  getSlice(): number;
  setSlice(value: number): void;

  getLow(): number;
  setLow(value: number): void;

  getHigh(): number;
  setHigh(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SlotRange.AsObject;
  static toObject(includeInstance: boolean, msg: SlotRange): SlotRange.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SlotRange, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SlotRange;
  static deserializeBinaryFromReader(message: SlotRange, reader: jspb.BinaryReader): SlotRange;
}

export namespace SlotRange {
  export type AsObject = {
    slice: number,
    low: number,
    high: number,
  }
}

export class Drive extends jspb.Message {
  getMount(): string;
  setMount(value: string): void;

  getKind(): Drive.Kind;
  setKind(value: Drive.Kind): void;

  hasStats(): boolean;
  clearStats(): void;
  getStats(): DriveStats | undefined;
  setStats(value?: DriveStats): void;

  getWorking(): boolean;
  setWorking(value: boolean): void;

  getFilesystem(): string;
  setFilesystem(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Drive.AsObject;
  static toObject(includeInstance: boolean, msg: Drive): Drive.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Drive, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Drive;
  static deserializeBinaryFromReader(message: Drive, reader: jspb.BinaryReader): Drive;
}

export namespace Drive {
  export type AsObject = {
    mount: string,
    kind: Drive.Kind,
    stats?: DriveStats.AsObject,
    working: boolean,
    filesystem: string,
  }

  export enum Kind {
    HDD = 0,
    SSD = 1,
    NVME = 2,
  }
}

export class DriveStats extends jspb.Message {
  getSize(): number;
  setSize(value: number): void;

  getUsed(): number;
  setUsed(value: number): void;

  getAvail(): number;
  setAvail(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DriveStats.AsObject;
  static toObject(includeInstance: boolean, msg: DriveStats): DriveStats.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DriveStats, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DriveStats;
  static deserializeBinaryFromReader(message: DriveStats, reader: jspb.BinaryReader): DriveStats;
}

export namespace DriveStats {
  export type AsObject = {
    size: number,
    used: number,
    avail: number,
  }
}

export class Bucket extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getAccesskey(): string;
  setAccesskey(value: string): void;

  getSecretkey(): string;
  setSecretkey(value: string): void;

  getUrl(): string;
  setUrl(value: string): void;

  getApi(): Bucket.API;
  setApi(value: Bucket.API): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Bucket.AsObject;
  static toObject(includeInstance: boolean, msg: Bucket): Bucket.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Bucket, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Bucket;
  static deserializeBinaryFromReader(message: Bucket, reader: jspb.BinaryReader): Bucket;
}

export namespace Bucket {
  export type AsObject = {
    id: string,
    accesskey: string,
    secretkey: string,
    url: string,
    api: Bucket.API,
  }

  export enum API {
    S3 = 0,
  }
}

export class RecordID extends jspb.Message {
  getEpoch(): number;
  setEpoch(value: number): void;

  getSeq(): number;
  setSeq(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RecordID.AsObject;
  static toObject(includeInstance: boolean, msg: RecordID): RecordID.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RecordID, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RecordID;
  static deserializeBinaryFromReader(message: RecordID, reader: jspb.BinaryReader): RecordID;
}

export namespace RecordID {
  export type AsObject = {
    epoch: number,
    seq: number,
  }
}

export class Record extends jspb.Message {
  hasKey(): boolean;
  clearKey(): void;
  getKey(): Projection | undefined;
  setKey(value?: Projection): void;

  hasSlice(): boolean;
  clearSlice(): void;
  getSlice(): Projection | undefined;
  setSlice(value?: Projection): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Record.AsObject;
  static toObject(includeInstance: boolean, msg: Record): Record.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Record, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Record;
  static deserializeBinaryFromReader(message: Record, reader: jspb.BinaryReader): Record;
}

export namespace Record {
  export type AsObject = {
    key?: Projection.AsObject,
    slice?: Projection.AsObject,
  }
}

export class Projection extends jspb.Message {
  getCodec(): Codec;
  setCodec(value: Codec): void;

  clearNamesList(): void;
  getNamesList(): Array<string>;
  setNamesList(value: Array<string>): void;
  addNames(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Projection.AsObject;
  static toObject(includeInstance: boolean, msg: Projection): Projection.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Projection, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Projection;
  static deserializeBinaryFromReader(message: Projection, reader: jspb.BinaryReader): Projection;
}

export namespace Projection {
  export type AsObject = {
    codec: Codec,
    namesList: Array<string>,
  }

  export class Field extends jspb.Message {
    getId(): number;
    setId(value: number): void;

    getName(): string;
    setName(value: string): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Field.AsObject;
    static toObject(includeInstance: boolean, msg: Field): Field.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Field, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Field;
    static deserializeBinaryFromReader(message: Field, reader: jspb.BinaryReader): Field;
  }

  export namespace Field {
    export type AsObject = {
      id: number,
      name: string,
    }
  }
}

export class Index extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Index.AsObject;
  static toObject(includeInstance: boolean, msg: Index): Index.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Index, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Index;
  static deserializeBinaryFromReader(message: Index, reader: jspb.BinaryReader): Index;
}

export namespace Index {
  export type AsObject = {
  }

  export enum Type {
    BTREE = 0,
    PREFIX = 2,
    RTREE = 3,
    FULLTEXT = 1,
  }
}

export class Topic extends jspb.Message {
  getSchema(): string;
  setSchema(value: string): void;

  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getSlot(): number;
  setSlot(value: number): void;

  getType(): Topic.Type;
  setType(value: Topic.Type): void;

  getQueueid(): number;
  setQueueid(value: number): void;

  getRollerid(): string;
  setRollerid(value: string): void;

  getMode(): Topic.Mode;
  setMode(value: Topic.Mode): void;

  getWritespeed(): number;
  setWritespeed(value: number): void;

  getCodec(): Codec;
  setCodec(value: Codec): void;

  hasKey(): boolean;
  clearKey(): void;
  getKey(): Projection | undefined;
  setKey(value?: Projection): void;

  hasSlicekey(): boolean;
  clearSlicekey(): void;
  getSlicekey(): Projection | undefined;
  setSlicekey(value?: Projection): void;

  getDrive(): Drive.Kind;
  setDrive(value: Drive.Kind): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Topic.AsObject;
  static toObject(includeInstance: boolean, msg: Topic): Topic.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Topic, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Topic;
  static deserializeBinaryFromReader(message: Topic, reader: jspb.BinaryReader): Topic;
}

export namespace Topic {
  export type AsObject = {
    schema: string,
    id: number,
    name: string,
    slot: number,
    type: Topic.Type,
    queueid: number,
    rollerid: string,
    mode: Topic.Mode,
    writespeed: number,
    codec: Codec,
    key?: Projection.AsObject,
    slicekey?: Projection.AsObject,
    drive: Drive.Kind,
  }

  export enum Type {
    STANDARD = 0,
  }

  export enum Mode {
    LOG = 0,
    QUEUE = 1,
    TABLE = 2,
    CACHE = 3,
  }
}

export class Roller extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getMinbytes(): number;
  setMinbytes(value: number): void;

  getMinage(): number;
  setMinage(value: number): void;

  getMincount(): number;
  setMincount(value: number): void;

  getMaxbytes(): number;
  setMaxbytes(value: number): void;

  getMaxage(): number;
  setMaxage(value: number): void;

  getMaxcount(): number;
  setMaxcount(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Roller.AsObject;
  static toObject(includeInstance: boolean, msg: Roller): Roller.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Roller, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Roller;
  static deserializeBinaryFromReader(message: Roller, reader: jspb.BinaryReader): Roller;
}

export namespace Roller {
  export type AsObject = {
    id: number,
    name: string,
    minbytes: number,
    minage: number,
    mincount: number,
    maxbytes: number,
    maxage: number,
    maxcount: number,
  }
}

export class Path extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getVolumeid(): string;
  setVolumeid(value: string): void;

  getDrive(): string;
  setDrive(value: string): void;

  getLocal(): boolean;
  setLocal(value: boolean): void;

  getBucket(): boolean;
  setBucket(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Path.AsObject;
  static toObject(includeInstance: boolean, msg: Path): Path.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Path, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Path;
  static deserializeBinaryFromReader(message: Path, reader: jspb.BinaryReader): Path;
}

export namespace Path {
  export type AsObject = {
    name: string,
    volumeid: string,
    drive: string,
    local: boolean,
    bucket: boolean,
  }

  export enum Type {
    LOCAL = 0,
    BUCKET = 1,
  }
}

export class Hash extends jspb.Message {
  getAlgorithm(): string;
  setAlgorithm(value: string): void;

  getValue(): Uint8Array | string;
  getValue_asU8(): Uint8Array;
  getValue_asB64(): string;
  setValue(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Hash.AsObject;
  static toObject(includeInstance: boolean, msg: Hash): Hash.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Hash, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Hash;
  static deserializeBinaryFromReader(message: Hash, reader: jspb.BinaryReader): Hash;
}

export namespace Hash {
  export type AsObject = {
    algorithm: string,
    value: Uint8Array | string,
  }

  export enum Algorithm {
    CRC32 = 0,
  }
}

export class Segment extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getTopicid(): number;
  setTopicid(value: number): void;

  getSlice(): number;
  setSlice(value: number): void;

  hasPath(): boolean;
  clearPath(): void;
  getPath(): Path | undefined;
  setPath(value?: Path): void;

  hasHeader(): boolean;
  clearHeader(): void;
  getHeader(): SegmentHeader | undefined;
  setHeader(value?: SegmentHeader): void;

  hasStats(): boolean;
  clearStats(): void;
  getStats(): SegmentStats | undefined;
  setStats(value?: SegmentStats): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Segment.AsObject;
  static toObject(includeInstance: boolean, msg: Segment): Segment.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Segment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Segment;
  static deserializeBinaryFromReader(message: Segment, reader: jspb.BinaryReader): Segment;
}

export namespace Segment {
  export type AsObject = {
    id: number,
    topicid: number,
    slice: number,
    path?: Path.AsObject,
    header?: SegmentHeader.AsObject,
    stats?: SegmentStats.AsObject,
  }
}

export class SegmentStats extends jspb.Message {
  hasHash(): boolean;
  clearHash(): void;
  getHash(): Hash | undefined;
  setHash(value?: Hash): void;

  getCount(): number;
  setCount(value: number): void;

  getHeader(): number;
  setHeader(value: number): void;

  getBody(): number;
  setBody(value: number): void;

  getSize(): number;
  setSize(value: number): void;

  getMaxbody(): number;
  setMaxbody(value: number): void;

  hasFirst(): boolean;
  clearFirst(): void;
  getFirst(): RecordPointer | undefined;
  setFirst(value?: RecordPointer): void;

  hasLast(): boolean;
  clearLast(): void;
  getLast(): RecordPointer | undefined;
  setLast(value?: RecordPointer): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SegmentStats.AsObject;
  static toObject(includeInstance: boolean, msg: SegmentStats): SegmentStats.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SegmentStats, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SegmentStats;
  static deserializeBinaryFromReader(message: SegmentStats, reader: jspb.BinaryReader): SegmentStats;
}

export namespace SegmentStats {
  export type AsObject = {
    hash?: Hash.AsObject,
    count: number,
    header: number,
    body: number,
    size: number,
    maxbody: number,
    first?: RecordPointer.AsObject,
    last?: RecordPointer.AsObject,
  }
}

export class SegmentHeader extends jspb.Message {
  getTimestamp(): number;
  setTimestamp(value: number): void;

  getTopicid(): number;
  setTopicid(value: number): void;

  getLogid(): number;
  setLogid(value: number): void;

  getStartindex(): number;
  setStartindex(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SegmentHeader.AsObject;
  static toObject(includeInstance: boolean, msg: SegmentHeader): SegmentHeader.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SegmentHeader, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SegmentHeader;
  static deserializeBinaryFromReader(message: SegmentHeader, reader: jspb.BinaryReader): SegmentHeader;
}

export namespace SegmentHeader {
  export type AsObject = {
    timestamp: number,
    topicid: number,
    logid: number,
    startindex: number,
  }
}

export class GlobalID extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getSlot(): number;
  setSlot(value: number): void;

  hasRecid(): boolean;
  clearRecid(): void;
  getRecid(): RecordID | undefined;
  setRecid(value?: RecordID): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GlobalID.AsObject;
  static toObject(includeInstance: boolean, msg: GlobalID): GlobalID.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GlobalID, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GlobalID;
  static deserializeBinaryFromReader(message: GlobalID, reader: jspb.BinaryReader): GlobalID;
}

export namespace GlobalID {
  export type AsObject = {
    id: number,
    slot: number,
    recid?: RecordID.AsObject,
  }
}

export class RecordPointer extends jspb.Message {
  hasId(): boolean;
  clearId(): void;
  getId(): RecordID | undefined;
  setId(value?: RecordID): void;

  getLogid(): number;
  setLogid(value: number): void;

  getSlot(): number;
  setSlot(value: number): void;

  getSize(): number;
  setSize(value: number): void;

  getPos(): number;
  setPos(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RecordPointer.AsObject;
  static toObject(includeInstance: boolean, msg: RecordPointer): RecordPointer.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RecordPointer, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RecordPointer;
  static deserializeBinaryFromReader(message: RecordPointer, reader: jspb.BinaryReader): RecordPointer;
}

export namespace RecordPointer {
  export type AsObject = {
    id?: RecordID.AsObject,
    logid: number,
    slot: number,
    size: number,
    pos: number,
  }
}

export class Queue extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getRequestid(): number;
  setRequestid(value: number): void;

  getReplyid(): number;
  setReplyid(value: number): void;

  getErrorid(): number;
  setErrorid(value: number): void;

  getLevel(): Level;
  setLevel(value: Level): void;

  getFifo(): boolean;
  setFifo(value: boolean): void;

  getMaxinflight(): number;
  setMaxinflight(value: number): void;

  getMaxvisibility(): number;
  setMaxvisibility(value: number): void;

  getMaxdelay(): number;
  setMaxdelay(value: number): void;

  getMaxretries(): number;
  setMaxretries(value: number): void;

  getAppid(): string;
  setAppid(value: string): void;

  clearTagsList(): void;
  getTagsList(): Array<string>;
  setTagsList(value: Array<string>): void;
  addTags(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Queue.AsObject;
  static toObject(includeInstance: boolean, msg: Queue): Queue.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Queue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Queue;
  static deserializeBinaryFromReader(message: Queue, reader: jspb.BinaryReader): Queue;
}

export namespace Queue {
  export type AsObject = {
    id: number,
    name: string,
    requestid: number,
    replyid: number,
    errorid: number,
    level: Level,
    fifo: boolean,
    maxinflight: number,
    maxvisibility: number,
    maxdelay: number,
    maxretries: number,
    appid: string,
    tagsList: Array<string>,
  }
}

export class Daemon extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getLevel(): Level;
  setLevel(value: Level): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Daemon.AsObject;
  static toObject(includeInstance: boolean, msg: Daemon): Daemon.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Daemon, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Daemon;
  static deserializeBinaryFromReader(message: Daemon, reader: jspb.BinaryReader): Daemon;
}

export namespace Daemon {
  export type AsObject = {
    id: number,
    name: string,
    level: Level,
  }
}

export class InitNode extends jspb.Message {
  hasNode(): boolean;
  clearNode(): void;
  getNode(): Node | undefined;
  setNode(value?: Node): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): InitNode.AsObject;
  static toObject(includeInstance: boolean, msg: InitNode): InitNode.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: InitNode, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): InitNode;
  static deserializeBinaryFromReader(message: InitNode, reader: jspb.BinaryReader): InitNode;
}

export namespace InitNode {
  export type AsObject = {
    node?: Node.AsObject,
  }
}

export class AddNodeToGroup extends jspb.Message {
  getNodeid(): string;
  setNodeid(value: string): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getSuffrage(): Suffrage;
  setSuffrage(value: Suffrage): void;

  getBootstrap(): boolean;
  setBootstrap(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddNodeToGroup.AsObject;
  static toObject(includeInstance: boolean, msg: AddNodeToGroup): AddNodeToGroup.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddNodeToGroup, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddNodeToGroup;
  static deserializeBinaryFromReader(message: AddNodeToGroup, reader: jspb.BinaryReader): AddNodeToGroup;
}

export namespace AddNodeToGroup {
  export type AsObject = {
    nodeid: string,
    groupid: number,
    suffrage: Suffrage,
    bootstrap: boolean,
  }
}

export class CreateDatabaseRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateDatabaseRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateDatabaseRequest): CreateDatabaseRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateDatabaseRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateDatabaseRequest;
  static deserializeBinaryFromReader(message: CreateDatabaseRequest, reader: jspb.BinaryReader): CreateDatabaseRequest;
}

export namespace CreateDatabaseRequest {
  export type AsObject = {
    name: string,
  }
}

export class CreateDatabaseReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateDatabaseReply.AsObject;
  static toObject(includeInstance: boolean, msg: CreateDatabaseReply): CreateDatabaseReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateDatabaseReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateDatabaseReply;
  static deserializeBinaryFromReader(message: CreateDatabaseReply, reader: jspb.BinaryReader): CreateDatabaseReply;
}

export namespace CreateDatabaseReply {
  export type AsObject = {
  }
}

export class TxCreateTopic extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getLevel(): Level;
  setLevel(value: Level): void;

  getRoller(): string;
  setRoller(value: string): void;

  getAppid(): string;
  setAppid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxCreateTopic.AsObject;
  static toObject(includeInstance: boolean, msg: TxCreateTopic): TxCreateTopic.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxCreateTopic, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxCreateTopic;
  static deserializeBinaryFromReader(message: TxCreateTopic, reader: jspb.BinaryReader): TxCreateTopic;
}

export namespace TxCreateTopic {
  export type AsObject = {
    name: string,
    level: Level,
    roller: string,
    appid: string,
  }
}

export class TxCreateQueue extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getLevel(): Level;
  setLevel(value: Level): void;

  getRoller(): string;
  setRoller(value: string): void;

  getFifo(): boolean;
  setFifo(value: boolean): void;

  getMaxinflight(): number;
  setMaxinflight(value: number): void;

  getMaxvisibility(): number;
  setMaxvisibility(value: number): void;

  getMaxdelay(): number;
  setMaxdelay(value: number): void;

  getMaxretries(): number;
  setMaxretries(value: number): void;

  getAppid(): string;
  setAppid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxCreateQueue.AsObject;
  static toObject(includeInstance: boolean, msg: TxCreateQueue): TxCreateQueue.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxCreateQueue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxCreateQueue;
  static deserializeBinaryFromReader(message: TxCreateQueue, reader: jspb.BinaryReader): TxCreateQueue;
}

export namespace TxCreateQueue {
  export type AsObject = {
    name: string,
    level: Level,
    roller: string,
    fifo: boolean,
    maxinflight: number,
    maxvisibility: number,
    maxdelay: number,
    maxretries: number,
    appid: string,
  }
}

export class TxCreateSegment extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxCreateSegment.AsObject;
  static toObject(includeInstance: boolean, msg: TxCreateSegment): TxCreateSegment.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxCreateSegment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxCreateSegment;
  static deserializeBinaryFromReader(message: TxCreateSegment, reader: jspb.BinaryReader): TxCreateSegment;
}

export namespace TxCreateSegment {
  export type AsObject = {
  }
}

export class TxRoll extends jspb.Message {
  getRollerid(): number;
  setRollerid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxRoll.AsObject;
  static toObject(includeInstance: boolean, msg: TxRoll): TxRoll.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxRoll, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxRoll;
  static deserializeBinaryFromReader(message: TxRoll, reader: jspb.BinaryReader): TxRoll;
}

export namespace TxRoll {
  export type AsObject = {
    rollerid: number,
  }
}

export class TxDeleteTopic extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxDeleteTopic.AsObject;
  static toObject(includeInstance: boolean, msg: TxDeleteTopic): TxDeleteTopic.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxDeleteTopic, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxDeleteTopic;
  static deserializeBinaryFromReader(message: TxDeleteTopic, reader: jspb.BinaryReader): TxDeleteTopic;
}

export namespace TxDeleteTopic {
  export type AsObject = {
  }
}

export class TxChangeRing extends jspb.Message {
  clearFromList(): void;
  getFromList(): Array<Slice>;
  setFromList(value: Array<Slice>): void;
  addFrom(value?: Slice, index?: number): Slice;

  clearToList(): void;
  getToList(): Array<Slice>;
  setToList(value: Array<Slice>): void;
  addTo(value?: Slice, index?: number): Slice;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxChangeRing.AsObject;
  static toObject(includeInstance: boolean, msg: TxChangeRing): TxChangeRing.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxChangeRing, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxChangeRing;
  static deserializeBinaryFromReader(message: TxChangeRing, reader: jspb.BinaryReader): TxChangeRing;
}

export namespace TxChangeRing {
  export type AsObject = {
    fromList: Array<Slice.AsObject>,
    toList: Array<Slice.AsObject>,
  }
}

export class TxChangeRingCancel extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxChangeRingCancel.AsObject;
  static toObject(includeInstance: boolean, msg: TxChangeRingCancel): TxChangeRingCancel.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxChangeRingCancel, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxChangeRingCancel;
  static deserializeBinaryFromReader(message: TxChangeRingCancel, reader: jspb.BinaryReader): TxChangeRingCancel;
}

export namespace TxChangeRingCancel {
  export type AsObject = {
  }
}

export class TxSplitTopic extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TxSplitTopic.AsObject;
  static toObject(includeInstance: boolean, msg: TxSplitTopic): TxSplitTopic.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TxSplitTopic, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TxSplitTopic;
  static deserializeBinaryFromReader(message: TxSplitTopic, reader: jspb.BinaryReader): TxSplitTopic;
}

export namespace TxSplitTopic {
  export type AsObject = {
  }
}

export enum Level {
  MISSION = 0,
  BUSINESS = 1,
  BACKGROUND = 2,
}

export enum Codec {
  JSON = 0,
  PROTOBUF = 1,
  MSGPACK = 2,
  CBOR = 3,
}

export enum Suffrage {
  VOTER = 0,
  NON_VOTER = 1,
  STAGING = 2,
}

export enum RaftStatus {
  FOLLOWER = 0,
  CANDIDATE = 1,
  LEADER = 2,
  SHUTDOWN = 3,
}

