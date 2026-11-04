import { create } from 'zustand';
export type User = {id:number;email:string;name:string;role:string};
type AuthState = {user:User|null;setUser:(user:User|null)=>void;logout:()=>void};
export const useAuth = create<AuthState>((set)=>({user:null,setUser:(user)=>set({user}),logout:()=>{localStorage.removeItem('taskflow_access');localStorage.removeItem('taskflow_refresh');set({user:null});}}));
